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

// Resources リソースのリスト
type Resources []Resource

type ResourceWalkFunc func(Resource) error

// Walk 各リソースに対し順次fnを適用する
//
// fnの適用は深さ優先(DFS)で行われる
// fnがerrorを返した場合は即時リターンし以降のリソースに対する処理は行われない
func (r *Resources) Walk(fn ResourceWalkFunc) error {
	return r.walk(*r, fn)
}

func (r *Resources) walk(targets Resources, fn ResourceWalkFunc) error {
	for _, target := range targets {
		if err := fn(target); err != nil {
			return err
		}
		for _, child := range target.Resources() {
			if err := r.walk(child.Resources(), fn); err != nil {
				return err
			}
		}
	}
	return nil
}
