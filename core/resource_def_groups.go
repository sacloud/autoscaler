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

	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ResourceDefGroups 一意な名前をキーとするリソースのリスト
type ResourceDefGroups struct {
	groups map[string]*ResourceDefGroup
}

func newResourceDefGroups() *ResourceDefGroups {
	return &ResourceDefGroups{
		groups: make(map[string]*ResourceDefGroup),
	}
}

func (rg *ResourceDefGroups) Get(key string) *ResourceDefGroup {
	v, _ := rg.GetOk(key)
	return v
}

func (rg *ResourceDefGroups) GetOk(key string) (*ResourceDefGroup, bool) {
	v, ok := rg.groups[key]
	return v, ok
}

func (rg *ResourceDefGroups) All() []*ResourceDefGroup {
	var values []*ResourceDefGroup
	for _, v := range rg.groups {
		values = append(values, v)
	}
	return values
}

func (rg *ResourceDefGroups) Set(key string, group *ResourceDefGroup) {
	group.name = key
	rg.groups[key] = group
}

func (rg *ResourceDefGroups) UnmarshalYAML(data []byte) error {
	var loaded map[string]*ResourceDefGroup
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return err
	}
	for k, v := range loaded {
		v.name = k
	}
	*rg = ResourceDefGroups{groups: loaded}
	return nil
}

func (rg *ResourceDefGroups) Validate(ctx context.Context, apiClient sacloud.APICaller, handlers Handlers) []error {
	var errors []error
	for _, group := range rg.groups {
		if errs := group.Validate(ctx, apiClient, handlers); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}
	return errors
}
