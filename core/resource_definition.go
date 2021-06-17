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

// ResourceDefinition Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
type ResourceDefinition interface {
	Type() ResourceTypes // リソースの型
	Selector() *ResourceSelector
	Validate(ctx context.Context, apiClient sacloud.APICaller) []error

	// Compute 現在/あるべき姿を算出する
	//
	// TypeとSelectorを元にさくらのクラウドAPIを用いて実リソースを検索、Resourceを作成して返す
	Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error)

	// Children このリソースに対する子リソースを返す
	Children() ResourceDefinitions
}

type ChildResourceDefinition interface {
	Parent() ResourceDefinition
	SetParent(parent ResourceDefinition)
}

// ResourceDefBase 全てのリソース定義が実装すべき基本プロパティ
//
// Resourceの実装に埋め込む場合、Compute()でComputedCacheを設定すること
type ResourceDefBase struct {
	TypeName       string              `yaml:"type"`
	TargetSelector *ResourceSelector   `yaml:"selector"`
	children       ResourceDefinitions `yaml:"-"`
}

func (r *ResourceDefBase) Type() ResourceTypes {
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

func (r *ResourceDefBase) Selector() *ResourceSelector {
	return r.TargetSelector
}

// Children 子リソースを返す(自身は含まない)
func (r *ResourceDefBase) Children() ResourceDefinitions {
	return r.children
}
