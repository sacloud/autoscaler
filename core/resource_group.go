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
	"fmt"
	"log"

	"github.com/sacloud/autoscaler/defaults"

	"github.com/sacloud/autoscaler/request"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceGroup struct {
	Name string `yaml:"-"` // ResourceGroupsのアンマーシャル時に設定される

	Actions   Actions   `yaml:"actions"`
	Resources Resources `yaml:"resources"`
}

func (rg *ResourceGroup) UnmarshalYAML(data []byte) error {
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	resourceGroup := &ResourceGroup{}
	resources := rawMap["resources"].([]interface{})
	for _, rawResource := range resources {
		v := rawResource.(map[string]interface{})
		resource, err := rg.unmarshalResourceFromMap(v)
		if err != nil {
			return err
		}

		rg.setParentResource(nil, resource)
		resourceGroup.Resources = append(resourceGroup.Resources, resource)
	}

	if rawActions, ok := rawMap["actions"]; ok {
		resourceGroup.Actions = Actions{}

		actions := rawActions.(map[string]interface{})
		for k, v := range actions {
			var handlers []string

			v := v.([]interface{})
			for _, v := range v {
				if v, ok := v.(string); ok {
					handlers = append(handlers, v)
				}
			}

			resourceGroup.Actions[k] = handlers
		}
	}

	*rg = *resourceGroup
	return nil
}

func (rg *ResourceGroup) unmarshalResourceFromMap(data map[string]interface{}) (Resource, error) {
	rawTypeName, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("'type' field required: %v", data)
	}
	typeName, ok := rawTypeName.(string)
	if !ok {
		return nil, fmt.Errorf("'type' is not string: %v", data)
	}

	remarshelded, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("yaml.Marshal failed with %v", data)
	}

	var resources Resources
	if rawChildren, ok := data["resources"]; ok {
		if children, ok := rawChildren.([]interface{}); ok {
			for _, child := range children {
				if c, ok := child.(map[string]interface{}); ok {
					r, err := rg.unmarshalResourceFromMap(c)
					if err != nil {
						return nil, err
					}
					resources = append(resources, r)
				}
			}
		}
	}

	var resource Resource
	switch typeName {
	case "Server":
		v := &Server{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "ServerGroup":
		v := &ServerGroup{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "EnhancedLoadBalancer", "ELB":
		v := &EnhancedLoadBalancer{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		// TypeNameのエイリアスを正規化
		v.TypeName = "EnhancedLoadBalancer"
		v.Children = resources
		resource = v
	case "GSLB":
		v := &GSLB{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "DNS":
		v := &DNS{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "Router":
		v := &Router{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	default:
		return nil, fmt.Errorf("received unexpected type: %s", typeName)
	}

	return resource, nil
}

func (rg *ResourceGroup) setParentResource(parent, r Resource) {
	if parent != nil {
		if v, ok := r.(ChildResource); ok {
			v.SetParent(parent)
		}
	}
	for _, child := range r.Resources() {
		rg.setParentResource(r, child)
	}
}

func (rg *ResourceGroup) ValidateActions(actionName string, handlerFilters Handlers) error {
	_, err := rg.handlers(actionName, handlerFilters)
	return err
}

func (rg *ResourceGroup) HandleAll(ctx *Context, apiClient sacloud.APICaller, handlerFilters Handlers) {
	job := ctx.Job()
	job.SetStatus(request.ScalingJobStatus_JOB_RUNNING)

	handlers, err := rg.handlers(ctx.Request().action, handlerFilters)
	if err != nil { // 事前にValidateHandlerFiltersで検証しておくため基本的にありえないはず
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		log.Printf("[FATAL] %s\n", err)
		return
	}

	if err := rg.handleAll(ctx, apiClient, handlers); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		log.Printf("[WARN] %s\n", err)
		return
	}

	job.SetStatus(request.ScalingJobStatus_JOB_DONE)
}

func (rg *ResourceGroup) handleAll(ctx *Context, apiClient sacloud.APICaller, handlers Handlers) error {
	forwardFn, backwardFn := rg.resourceWalkFuncs(ctx, apiClient, handlers)
	if err := rg.Resources.Walk(forwardFn, backwardFn); err != nil {
		return err
	}
	rg.clearCacheAll()
	return nil
}

func (rg *ResourceGroup) resourceWalkFuncs(parentCtx *Context, apiClient sacloud.APICaller, handlers Handlers) (ResourceWalkFunc, ResourceWalkFunc) {
	// TODO 並列化
	forwardFn := func(resource Resource) error {
		_, err := resource.Compute(parentCtx, apiClient)
		return err
	}

	backwardFn := func(resource Resource) error {
		computed := resource.Computed()

		ctx := NewHandlingContext(parentCtx, computed)

		// preHandle
		if err := rg.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
			return h.PreHandle(ctx, c)
		}); err != nil {
			return err
		}

		// handle
		if err := rg.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
			return h.Handle(ctx, c)
		}); err != nil {
			return err
		}

		// refresh
		refreshed, err := resource.Compute(ctx.ForRefresh(), apiClient)
		if err != nil {
			return err
		}
		computed = refreshed

		// postHandle
		if err := rg.handleAllByFunc(computed, handlers, func(h *Handler, c Computed) error {
			return h.PostHandle(ctx, c)
		}); err != nil {
			return err
		}

		return nil
	}
	return forwardFn, backwardFn
}

func (rg *ResourceGroup) handleAllByFunc(computed Computed, handlers Handlers, fn func(*Handler, Computed) error) error {
	for _, handler := range handlers {
		if err := fn(handler, computed); err != nil {
			return err
		}
	}
	return nil
}

func (rg *ResourceGroup) clearCacheAll() {
	rg.Resources.Walk(func(resource Resource) error { // nolint 戻り値のerrorを無視しているがerrorが返ることはない
		resource.ClearCache()
		return nil
	}, nil)
}

// Handlers 引数で指定されたハンドラーのリストをActionsの定義に合致するハンドラだけにフィルタして返す
func (rg *ResourceGroup) handlers(actionName string, allHandlers Handlers) (Handlers, error) {
	var results Handlers

	if len(rg.Actions) == 0 {
		for _, h := range allHandlers {
			if !h.Disabled {
				results = append(results, h)
			}
		}
		return results, nil
	}

	if actionName == "" || actionName == defaults.ActionName {
		// Actionsが定義されている時にActionNameが省略 or デフォルトの場合はActionsの最初のキーを利用
		// YAMLでの定義順とは限らないため注意
		for k := range rg.Actions {
			actionName = k
			break
		}
	}

	handlers, ok := rg.Actions[actionName]
	if !ok {
		return nil, fmt.Errorf("action %q not found", actionName)
	}
	if len(handlers) == 0 {
		return nil, fmt.Errorf("action %q is empty", actionName)
	}

	for _, name := range handlers {
		var found *Handler
		for _, h := range allHandlers {
			if h.Name == name {
				found = h
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("handler %q not found", name)
		}
		results = append(results, found)
	}
	return results, nil
}
