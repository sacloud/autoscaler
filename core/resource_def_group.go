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

	"github.com/sacloud/autoscaler/request"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ResourceDefGroup アクションとリソース定義の組み合わせ
type ResourceDefGroup struct {
	Actions      Actions             `yaml:"actions"`
	ResourceDefs ResourceDefinitions `yaml:"resources" validate:"required"`

	name string // ResourceGroupsのアンマーシャル時に設定される
}

func (rdg *ResourceDefGroup) ValidateActions(actionName string, handlerFilters Handlers) error {
	_, err := rdg.handlers(actionName, handlerFilters)
	return err
}

func (rdg *ResourceDefGroup) Validate(ctx context.Context, apiClient sacloud.APICaller, handlers Handlers) []error {
	errors := &multierror.Error{}

	// Actions
	errors = multierror.Append(errors, rdg.Actions.Validate(ctx, handlers)...)
	// Resources
	errors = multierror.Append(errors, rdg.ResourceDefs.Validate(ctx, apiClient)...)

	// set group name prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource def group=%s:", rdg.name)).(*multierror.Error)

	return errors.Errors
}

// Handlers 引数で指定されたハンドラーのリストをActionsの定義に合致するハンドラだけにフィルタして返す
func (rdg *ResourceDefGroup) handlers(actionName string, allHandlers Handlers) (Handlers, error) {
	var results Handlers

	if len(rdg.Actions) == 0 {
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
		for k := range rdg.Actions {
			actionName = k
			break
		}
	}

	handlers, ok := rdg.Actions[actionName]
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

func (rdg *ResourceDefGroup) HandleAll(ctx *RequestContext, apiClient sacloud.APICaller, handlerFilters Handlers) {
	job := ctx.Job()
	job.SetStatus(request.ScalingJobStatus_JOB_RUNNING)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_RUNNING) // nolint

	handlers, err := rdg.handlers(ctx.Request().action, handlerFilters)
	if err != nil { // 事前にValidateHandlerFiltersで検証しておくため基本的にありえないはず
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		ctx.Logger().Warn("status", request.ScalingJobStatus_JOB_FAILED, "fatal", err) // nolint
	}

	if err := rdg.ResourceDefs.HandleAll(ctx, apiClient, handlers); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		ctx.Logger().Warn("status", request.ScalingJobStatus_JOB_FAILED, "error", err) // nolint
		return
	}

	job.SetStatus(request.ScalingJobStatus_JOB_DONE)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_DONE) // nolint
}
