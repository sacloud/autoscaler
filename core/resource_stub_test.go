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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type stubResource struct {
	*ResourceBase `yaml:",inline"`
	computeFunc   func(ctx *Context, apiClient sacloud.APICaller) (Computed, error)
}

func (r *stubResource) Validate() error {
	return nil
}

func (r *stubResource) Compute(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
	if r.computeFunc != nil {
		computed, err := r.computeFunc(ctx, apiClient)
		r.ComputedCache = computed
		return computed, err
	}
	return nil, nil
}

type stubComputed struct {
	id          string
	instruction handler.ResourceInstructions
	current     *handler.Resource
	desired     *handler.Resource
}

func (c *stubComputed) ID() string {
	return c.id
}

func (c *stubComputed) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *stubComputed) Current() *handler.Resource {
	return c.current
}

func (c *stubComputed) Desired() *handler.Resource {
	return c.desired
}
