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

	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDefServerGroup struct {
	*ResourceBase `yaml:",inline"`
	wrapper       ResourceDefinition
}

func (s *ResourceDefServerGroup) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	// TODO 実装
	return nil
}

func (s *ResourceDefServerGroup) Fetch(ctx *RequestContext, apiClient sacloud.APICaller) ([]Resource, error) {
	// TODO Fetchを実装する
	return nil, nil
}

func (s *ResourceDefServerGroup) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error) {
	// TODO 実装
	return nil, nil
}

// Parent ChildResourceインターフェースの実装
func (s *ResourceDefServerGroup) Parent() ResourceDefinition {
	return s.wrapper
}

// SetParent ChildResourceインターフェースの実装
func (s *ResourceDefServerGroup) SetParent(parent ResourceDefinition) {
	s.wrapper = parent
}
