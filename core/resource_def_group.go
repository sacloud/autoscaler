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
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ResourceDefGroup アクションとリソース定義の組み合わせ
type ResourceDefGroup struct {
	ResourceDefs ResourceDefinitions `yaml:"resources" validate:"required"`

	name string // ResourceGroupsのアンマーシャル時に設定される
}

func (rdg *ResourceDefGroup) Validate(ctx context.Context, apiClient sacloud.APICaller, handlers Handlers) []error {
	errors := &multierror.Error{}

	// Resources
	errors = multierror.Append(errors, rdg.ResourceDefs.Validate(ctx, apiClient)...)

	// set group name prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource def group=%s:", rdg.name)).(*multierror.Error)

	return errors.Errors
}

func (rdg *ResourceDefGroup) HandleAll(ctx *RequestContext, apiClient sacloud.APICaller, allHandlers Handlers) {
	job := ctx.Job()
	job.SetStatus(request.ScalingJobStatus_JOB_RUNNING)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_RUNNING) // nolint

	if err := rdg.ResourceDefs.HandleAll(ctx, apiClient, allHandlers); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_FAILED)
		ctx.Logger().Warn("status", request.ScalingJobStatus_JOB_FAILED, "error", err) // nolint
		return
	}

	job.SetStatus(request.ScalingJobStatus_JOB_DONE)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_DONE) // nolint
}
