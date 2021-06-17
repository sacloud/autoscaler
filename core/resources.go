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

// Resources リソースのリスト
type Resources []Resource

func (r *Resources) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	var errors []error

	fn := func(r Resource) error {
		if errs := r.Validate(ctx, apiClient); len(errs) > 0 {
			errors = append(errors, errs...)
		}
		return nil
	}

	r.Walk(fn, nil) // nolint
	return errors
}

type ResourceWalkFunc func(Resource) error

// Walk 各リソースに対し順次forwardFn,backwardFnを適用する
//
// forwardFnの適用は上から行われる
// backwardFnは末端から行われる
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
//    - backwardFn(resource3)
//    - forwardFn(resource4)
//    - backwardFn(resource4)
//    - backwardFn(resource2)
//    - backwardFn(resource1)
//
// forwardFn, backwardFnがerrorを返した場合は即時リターンし以降のリソースに対する処理は行われない
func (r *Resources) Walk(forwardFn, backwardFn ResourceWalkFunc) error {
	return r.walk(*r, forwardFn, backwardFn)
}

func (r *Resources) walk(targets Resources, forwardFn, backwardFn ResourceWalkFunc) error {
	noopFunc := func(_ Resource) error {
		return nil
	}
	if forwardFn == nil {
		forwardFn = noopFunc
	}
	if backwardFn == nil {
		backwardFn = noopFunc
	}

	for _, target := range targets {
		if err := forwardFn(target); err != nil {
			return err
		}
		for _, child := range target.Children() {
			if err := forwardFn(child); err != nil {
				return err
			}
			if err := r.walk(child.Children(), forwardFn, backwardFn); err != nil {
				return err
			}
			if err := backwardFn(child); err != nil {
				return err
			}
		}
		if err := backwardFn(target); err != nil {
			return err
		}
	}
	return nil
}
