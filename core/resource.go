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

// Resource Definitionから作られるResource
type Resource interface {
	// Compute リクエストに沿った、希望する状態を算出する
	//
	// refreshがtrueの場合、さくらのクラウドAPIを用いて最新の状態を取得した上で処理を行う
	// falseの場合はキャッシュされている結果を元に処理を行う
	Compute(ctx *RequestContext, refresh bool) (Computed, error)

	// Type リソースの型
	Type() ResourceTypes
	// Children 子リソース
	Children() Resources
	// AppendChildren 子リソースを設定
	AppendChildren(Resources)
	// Parent 親Resourceへの参照
	Parent() Resource
	// SetParent 親Resourceを設定
	SetParent(parent Resource)
}

// Resources Resourceのスライス
type Resources []Resource

// ResourceBase 全てのリソースが所有すべきResourceの基本構造
type ResourceBase struct {
	resourceType ResourceTypes
	parent       Resource
	children     Resources
}

func (r *ResourceBase) Type() ResourceTypes {
	return r.resourceType
}

func (r *ResourceBase) Children() Resources {
	return r.children
}

func (r *ResourceBase) AppendChildren(children Resources) {
	r.children = append(r.children, children...)
}

func (r *ResourceBase) Parent() Resource {
	return r.parent
}

func (r *ResourceBase) SetParent(parent Resource) {
	r.parent = parent
}
