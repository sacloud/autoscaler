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

import "fmt"

// Resource Definitionから作られるResource
type Resource interface {
	// Compute リクエストに沿った、希望する状態を算出する
	//
	// refreshがtrueの場合、さくらのクラウドAPIを用いて最新の状態を取得した上で処理を行う
	// falseの場合はキャッシュされている結果を元に処理を行う
	Compute(ctx *RequestContext, computedParent Computed, refresh bool) (Computed, error)

	// Type リソースの型
	Type() ResourceTypes

	// String Resourceの文字列表現
	String() string
}

// Resources Resourceのスライス
type Resources []Resource

func (rs *Resources) String() string {
	var types []string
	for _, r := range *rs {
		types = append(types, r.Type().String())
	}
	if len(types) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", types)
}

// ResourceBase 全てのリソースが所有すべきResourceの基本構造
type ResourceBase struct {
	resourceType ResourceTypes
}

func (r *ResourceBase) Type() ResourceTypes {
	return r.resourceType
}
