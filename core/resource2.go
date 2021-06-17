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

import "github.com/sacloud/libsacloud/v2/sacloud"

// Resource2 Definitionから作られるResource
//
// TODO 現行Resourceとの切り替え時に名前変更する
type Resource2 interface {
	// Type リソースの型
	Type() ResourceTypes

	// Compute リクエストに沿った、希望する状態を算出する
	//
	// refreshがtrueの場合、さくらのクラウドAPIを用いて最新の状態に更新した上で処理を行う
	Compute(ctx *RequestContext, apiClient sacloud.APICaller, refresh bool) (Computed, error)

	// Children 子リソース
	Children() Resources2
	// AppendChildren 子リソースを設定
	AppendChildren(Resources2)

	Parent() Resource2
}

type Resources2 []Resource2

type ResourceBase2 struct {
	resourceType ResourceTypes
	parent       Resource2
	children     Resources2
}

func NewResourceBase2(tp ResourceTypes, parent Resource2, children ...Resource2) *ResourceBase2 {
	v := &ResourceBase2{
		resourceType: tp,
		parent:       parent,
	}
	for _, child := range children {
		if child != nil {
			v.children = append(v.children, child)
		}
	}
	return v
}

func (r *ResourceBase2) Type() ResourceTypes {
	return r.resourceType
}

func (r *ResourceBase2) Children() Resources2 {
	return r.children
}

func (r *ResourceBase2) AppendChildren(children Resources2) {
	r.children = append(r.children, children...)
}

func (r *ResourceBase2) Parent() Resource2 {
	return r.parent
}
