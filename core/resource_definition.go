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
	"fmt"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ResourceDefinition Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
// Selectorを元にさくらのクラウドAPIで検索し、ヒットしたリソース全てがスケール操作対象となる
type ResourceDefinition interface {
	Type() ResourceTypes // リソースの型
	Selector() *ResourceSelector
	Validate(ctx context.Context, apiClient sacloud.APICaller) []error

	// Fetch TypeやSelectorの値からさくらのクラウド上のリソース情報をfetchし、Resourceを作成して返す
	Fetch(ctx *RequestContext, apiClient sacloud.APICaller) ([]Resource, error)

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
	Children() ResourceDefinitions
}

type ChildResource interface {
	Parent() ResourceDefinition
	SetParent(parent ResourceDefinition)
}

// ResourceBase 全てのリソースが実装すべき基本プロパティ
//
// Resourceの実装に埋め込む場合、Compute()でComputedCacheを設定すること
type ResourceBase struct {
	TypeName       string              `yaml:"type"`
	TargetSelector *ResourceSelector   `yaml:"selector"`
	children       ResourceDefinitions `yaml:"-"`
	ComputedCache  Computed            `yaml:"-"`
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

// Children 子リソースを返す(自身は含まない)
func (r *ResourceBase) Children() ResourceDefinitions {
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

// ResourceSelector さくらのクラウド上で対象リソースを特定するための情報を提供する
type ResourceSelector struct {
	ID    types.ID `yaml:"id"`
	Names []string `yaml:"names"`
	Zone  string   `yaml:"zone"` // グローバルリソースの場合はis1aまたは空とする
}

func (rs *ResourceSelector) String() string {
	if rs != nil {
		return fmt.Sprintf("ID: %s, Names: %s, Zone: %s", rs.ID, rs.Names, rs.Zone)
	}
	return ""
}

func (rs *ResourceSelector) findCondition() *sacloud.FindCondition {
	fc := &sacloud.FindCondition{
		Filter: search.Filter{},
	}
	if !rs.ID.IsEmpty() {
		fc.Filter[search.Key("ID")] = search.ExactMatch(rs.ID.String())
	}
	if len(rs.Names) != 0 {
		fc.Filter[search.Key("Name")] = search.PartialMatch(rs.Names...)
	}
	return fc
}
