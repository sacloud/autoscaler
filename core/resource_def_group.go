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

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ResourceDefGroup アクションとリソース定義の組み合わせ
type ResourceDefGroup struct {
	Actions      Actions             `yaml:"actions"`
	ResourceDefs ResourceDefinitions `yaml:"resources"`

	name string // ResourceGroupsのアンマーシャル時に設定される
}

// TODO UnmarshalYAMLを実装する

func (rg *ResourceDefGroup) ValidateActions(actionName string, handlerFilters Handlers) error {
	_, err := rg.handlers(actionName, handlerFilters)
	return err
}

func (rg *ResourceDefGroup) Validate(ctx context.Context, apiClient sacloud.APICaller, handlers Handlers) []error {
	errors := &multierror.Error{}

	// Actions
	errors = multierror.Append(errors, rg.Actions.Validate(ctx, handlers)...)
	// Resources
	errors = multierror.Append(errors, rg.ResourceDefs.Validate(ctx, apiClient)...)

	// set group name prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource def group=%s:", rg.name)).(*multierror.Error)

	return errors.Errors
}

func (rg *ResourceDefGroup) ResourceGroup(ctx *RequestContext, apiClient sacloud.APICaller) (*ResourceGroup2, error) {
	group := &ResourceGroup2{
		Resources: Resources2{},
	}

	if err := rg.buildResourceGroup(ctx, apiClient, group, nil, rg.ResourceDefs); err != nil {
		return nil, err
	}
	return group, nil
}

func (rg *ResourceDefGroup) buildResourceGroup(ctx *RequestContext, apiClient sacloud.APICaller, group *ResourceGroup2, parent Resource2, defs ResourceDefinitions) error {
	for _, def := range defs {
		resources, err := def.Compute(ctx, apiClient)
		if err != nil {
			return err
		}
		if len(resources) == 0 {
			return fmt.Errorf("ResourceDefinition did not return any resources")
		}
		if def.Children() != nil && len(resources) > 1 {
			// 親リソースになっている場合は複数リソースを許容しない
			return fmt.Errorf("ResourceDefinition with children must return one resource, but got multiple resources")
		}

		// 親リソースが指定されてたらそちらに、以外はgroupに直接追加
		if parent != nil {
			parent.SetChildren(append(parent.Children(), resources...))
		} else {
			group.Resources = append(group.Resources, resources...)
		}

		if err := rg.buildResourceGroup(ctx, apiClient, group, resources[0], def.Children()); err != nil {
			return err
		}
	}
	return nil
}

// Handlers 引数で指定されたハンドラーのリストをActionsの定義に合致するハンドラだけにフィルタして返す
func (rg *ResourceDefGroup) handlers(actionName string, allHandlers Handlers) (Handlers, error) {
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
