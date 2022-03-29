// Copyright 2021-2022 The sacloud/autoscaler Authors
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

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
)

// ResourceDefinitions リソースのリスト
type ResourceDefinitions []ResourceDefinition

func (rds *ResourceDefinitions) Validate(ctx context.Context, apiClient iaas.APICaller) []error {
	var errors []error
	names := make(map[string]struct{})

	fn := func(r ResourceDefinition) error {
		if err := validate.Struct(r); err != nil {
			errors = append(errors, multierror.Prefix(err, fmt.Sprintf("resource=%s", r.Type())))
		}
		if errs := r.Validate(ctx, apiClient); len(errs) > 0 {
			errors = append(errors, errs...)
		}
		if _, exist := names[r.Name()]; exist {
			errors = append(errors, fmt.Errorf("resource name %s is duplicated", r.Name()))
		}
		names[r.Name()] = struct{}{}
		return nil
	}

	if err := rds.walk(*rds, fn); err != nil {
		errors = append(errors, err)
	}
	return errors
}

func (rds *ResourceDefinitions) ResourceNames() []string {
	nameMap := make(map[string]struct{})
	fn := func(r ResourceDefinition) error {
		name := r.Name()
		if name != "" {
			nameMap[name] = struct{}{}
		}
		return nil
	}

	rds.walk(*rds, fn) // nolint

	var names []string
	for name := range nameMap {
		names = append(names, name)
	}
	return names
}

type resourceDefWalkFunc func(def ResourceDefinition) error

func (rds *ResourceDefinitions) walk(targets ResourceDefinitions, fn resourceDefWalkFunc) error {
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
	}
	return nil
}

func (rds *ResourceDefinitions) FilterByResourceName(name string) ResourceDefinitions {
	for _, r := range *rds {
		if r.Name() == name {
			return ResourceDefinitions{r}
		}
	}
	return nil
}

func (rds *ResourceDefinitions) HandleAll(ctx *RequestContext, apiClient iaas.APICaller, handlers Handlers) {
	job := ctx.Job()
	job.SetStatus(request.ScalingJobStatus_JOB_RUNNING)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_RUNNING) // nolint

	if err := rds.handleAll(ctx, apiClient, handlers, *rds); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		ctx.Logger().Warn("status", request.ScalingJobStatus_JOB_FAILED, "error", err) // nolint
		return
	}

	job.SetStatus(request.ScalingJobStatus_JOB_DONE)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_DONE) // nolint
}

func (rds *ResourceDefinitions) handleAll(ctx *RequestContext, apiClient iaas.APICaller, handlers Handlers, defs ResourceDefinitions) error {
	for _, def := range defs {
		resources, err := def.Compute(ctx, apiClient)
		if err != nil {
			return err
		}
		for _, resource := range resources {
			if err := rds.handleResource(ctx, handlers, resource); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rds *ResourceDefinitions) handleResource(parentCtx *RequestContext, handlers Handlers, resource Resource) error {
	computed, err := resource.Compute(parentCtx, false)
	if err != nil {
		return err
	}

	zone := computed.Zone()
	if zone == "" {
		zone = "global"
	}
	id := computed.ID()
	if id == "" {
		id = "(known after handle)"
	}
	handlingCtx := NewHandlingContext(parentCtx, computed).WithLogger("type", computed.Type(), "zone", zone, "id", id, "name", computed.Name())

	// preHandle
	if err := rds.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
		ctx := handlingCtx.WithLogger("step", "PreHandle", "handler", h.Name)
		if h.BuiltinHandler != nil {
			h.BuiltinHandler.SetLogger(ctx.Logger())
		}
		return h.PreHandle(ctx, c)
	}); err != nil {
		return err
	}

	// handle
	if err := rds.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
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
	// IDが採番されていたり変更されていたりするためHandlingContextも更新しておく
	id = refreshed.ID()
	if id == "" {
		id = "(known after handle)"
	}
	handlingCtx = NewHandlingContext(parentCtx, computed).WithLogger("type", refreshed.Type(), "zone", zone, "id", id, "name", refreshed.Name())
	computed = refreshed

	// postHandle
	if err := rds.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
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

func (rds *ResourceDefinitions) handleAllByFunc(computed Computed, handlers Handlers, fn func(*Handler, Computed) error) error {
	for _, handler := range handlers {
		if err := fn(handler, computed); err != nil {
			return err
		}
	}
	return nil
}
