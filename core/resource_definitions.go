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

type ResourceDefWalkFunc func(definitions ResourceDefinition) error

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
