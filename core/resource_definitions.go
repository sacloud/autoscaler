// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"context"
	"fmt"

	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ResourceDefinitions リソースのリスト
type ResourceDefinitions []ResourceDefinition

func (r *ResourceDefinitions) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	var errors []error

	fn := func(r ResourceDefinition) error {
		if errs := r.Validate(ctx, apiClient); len(errs) > 0 {
			errors = append(errors, errs...)
		}
		return nil
	}

	if err := r.Walk(fn); err != nil {
		errors = append(errors, err)
	}
	return errors
}

type ResourceDefWalkFunc func(def ResourceDefinition) error

// Walk 各リソースに対し順次fnを適用する
//
// forwardFnの適用は上から行われる
//
// example:
// resource1
//  |
//  |- resource2
//      |
//      |- resource3
//      |- resource4
//
//  この場合は以下の処理順になる
//    - forwardFn(resource1)
//    - forwardFn(resource2)
//    - forwardFn(resource3)
//    - forwardFn(resource4)
//
// fnがerrorを返した場合は即時リターンし以降のリソースに対する処理は行われない
func (r *ResourceDefinitions) Walk(fn ResourceDefWalkFunc) error {
	return r.walk(*r, fn)
}

func (r *ResourceDefinitions) walk(targets ResourceDefinitions, fn ResourceDefWalkFunc) error {
	noopFunc := func(_ ResourceDefinition) error {
		return nil
	}
	if fn == nil {
		fn = noopFunc
	}

	for _, target := range targets {
		if err := fn(target); err != nil {
			return err
		}
		for _, child := range target.Children() {
			if err := fn(child); err != nil {
				return err
			}
			if err := r.walk(child.Children(), fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ResourceDefinitions) HandleAll(ctx *RequestContext, apiClient sacloud.APICaller, handlers Handlers) error {
	return r.handleAll(ctx, apiClient, handlers, nil, *r)
}

func (r *ResourceDefinitions) handleAll(ctx *RequestContext, apiClient sacloud.APICaller, handlers Handlers, parentResource Resource, defs ResourceDefinitions) error {
	for _, def := range defs {
		resources, err := def.Compute(ctx, apiClient)
		if err != nil {
			return err
		}
		// 子リソースが定義されているリソースは複数ヒット時にはエラーとする
		children := def.Children()

		if len(children) > 0 && len(resources) > 1 {
			return fmt.Errorf("A resource definition with children cannot return multiple resources: definition: %#v, returned: %#v", def, resources)
		}

		for _, resource := range resources {
			if parentResource != nil {
				resource.SetParent(parentResource)
			}
			if len(children) > 0 {
				if err := r.handleAll(ctx, apiClient, handlers, resource, children); err != nil {
					return err
				}
			}

			if err := r.handleResource(ctx, handlers, resource); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ResourceDefinitions) handleResource(parentCtx *RequestContext, handlers Handlers, resource Resource) error {
	computed, err := resource.Compute(parentCtx, false)
	if err != nil {
		return err
	}

	zone := computed.Zone()
	if zone == "" {
		zone = "global"
	}
	handlingCtx := NewHandlingContext(parentCtx, computed).WithLogger("type", computed.Type(), "zone", zone, "id", computed.ID())

	// preHandle
	if err := r.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
		ctx := handlingCtx.WithLogger("step", "PreHandle", "handler", h.Name)
		if h.BuiltinHandler != nil {
			h.BuiltinHandler.SetLogger(ctx.Logger())
		}
		return h.PreHandle(ctx, c)
	}); err != nil {
		return err
	}

	// handle
	if err := r.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
		ctx := handlingCtx.WithLogger("step", "Handle", "handler", h.Name)
		if h.BuiltinHandler != nil {
			h.BuiltinHandler.SetLogger(ctx.Logger())
		}
		return h.Handle(ctx, c)
	}); err != nil {
		return err
	}

	// refresh
	refreshed, err := resource.Compute(handlingCtx.RequestContext, true)
	if err != nil {
		return err
	}
	computed = refreshed

	// postHandle
	if err := r.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
		ctx := handlingCtx.WithLogger("step", "PostHandle", "handler", h.Name)
		if h.BuiltinHandler != nil {
			h.BuiltinHandler.SetLogger(ctx.Logger())
		}
		return h.PostHandle(ctx, c)
	}); err != nil {
		return err
	}

	return nil
}

func (r *ResourceDefinitions) handleAllByFunc(computed Computed, handlers Handlers, fn func(*Handler, Computed) error) error {
	for _, handler := range handlers {
		if err := fn(handler, computed); err != nil {
			return err
		}
	}
	return nil
}
