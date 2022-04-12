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

	"github.com/sacloud/iaas-api-go"
)

type stubResourceDef struct {
	*ResourceDefBase
	computeFunc func(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error)

	Dummy        string `validate:"omitempty,oneof=value1 value2"`
	validateFunc func(ctx context.Context, apiClient iaas.APICaller) []error
}

func (d *stubResourceDef) String() string {
	return "stub"
}

func (d *stubResourceDef) Validate(ctx context.Context, apiClient iaas.APICaller) []error {
	if d.validateFunc != nil {
		return d.validateFunc(ctx, apiClient)
	}
	return nil
}

func (d *stubResourceDef) Compute(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error) {
	if d.computeFunc != nil {
		return d.computeFunc(ctx, apiClient)
	}
	return nil, nil
}

type stubResource struct {
	*ResourceBase
	computeFunc func(ctx *RequestContext, refresh bool) (Computed, error)
	name        string
	parent      Resource
}

func (r *stubResource) String() string {
	return r.name
}

func (r *stubResource) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if r.computeFunc != nil {
		return r.computeFunc(ctx, refresh)
	}
	return nil, nil
}

func (r *stubResource) Parent() Resource {
	return r.parent
}
