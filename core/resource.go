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

// Resource Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
type Resource interface {
	Type() ResourceTypes // リソースの型
	Selector() *ResourceSelector
	Validate(ctx context.Context, apiClient sacloud.APICaller) []error

	// Compute 現在/あるべき姿を算出する
	//
	// さくらのクラウド上の1つのリソースに対し1つのComputedを返す
	// Selector()の値によっては複数のComputedを返しても良い
	// Computeの結果はキャッシュしておき、Computed()で参照可能にしておく
	// キャッシュはClearCache()を呼ぶまで保持しておく
	Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error)

	// Computed Compute()の結果のキャッシュ、Compute()呼び出し前はnilを返す
	Computed() Computed

	// ClearCache Compute()の結果のキャッシュをクリアする
	ClearCache()

	// Children このリソースに対する子リソースを返す
	Children() Resources
}

type ChildResource interface {
	Parent() Resource
	SetParent(parent Resource)
}

// ResourceBase 全てのリソースが実装すべき基本プロパティ
//
// Resourceの実装に埋め込む場合、Compute()でComputedCacheを設定すること
type ResourceBase struct {
	TypeName       string            `yaml:"type"`
	TargetSelector *ResourceSelector `yaml:"selector"`
	children       Resources         `yaml:"-"`
	ComputedCache  Computed          `yaml:"-"`
}

func (r *ResourceBase) Type() ResourceTypes {
	switch r.TypeName {
	case ResourceTypeServer.String():
		return ResourceTypeServer
	case ResourceTypeServerGroup.String():
		return ResourceTypeServerGroup
	case ResourceTypeEnhancedLoadBalancer.String(), "ELB":
		return ResourceTypeEnhancedLoadBalancer
	case ResourceTypeGSLB.String():
		return ResourceTypeGSLB
	case ResourceTypeDNS.String():
		return ResourceTypeDNS
	}
	return ResourceTypeUnknown
}

func (r *ResourceBase) Selector() *ResourceSelector {
	return r.TargetSelector
}

// Resources 子リソースを返す(自身は含まない)
func (r *ResourceBase) Children() Resources {
	return r.children
}

// Computed 各リソースでのCompute()のキャッシュされた結果を返す
func (r *ResourceBase) Computed() Computed {
	return r.ComputedCache
}

// ClearCache Compute()の結果のキャッシュをクリア
func (r *ResourceBase) ClearCache() {
	r.ComputedCache = nil
}
