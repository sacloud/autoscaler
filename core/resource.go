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
	"fmt"

	"github.com/sacloud/libsacloud/v2/sacloud/search"

	"github.com/sacloud/libsacloud/v2/sacloud"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// Resource Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
type Resource interface {
	Type() ResourceTypes // リソースの型
	Selector() *ResourceSelector
	Desired(ctx *Context, apiClient sacloud.APICaller) (Desired, error)
	Validate() error
	Resources() Resources // 子リソース(GSLBに対する実サーバなど)
}

// ResourceBase 全てのリソースが実装すべき基本プロパティ
type ResourceBase struct {
	TypeName       string                   `yaml:"type"` // TODO enumにすべきか?
	TargetSelector *ResourceSelector        `yaml:"selector"`
	Children       Resources                `yaml:"wrappers"`
	TargetHandlers []*ResourceHandlerConfig `yaml:"handlers"`
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
	return ResourceTypeUnknown // TODO バリデーションなどで到達させないようにする
}

func (r *ResourceBase) Selector() *ResourceSelector {
	return r.TargetSelector
}

// Resources 子リソースを返す(自身は含まない)
func (r *ResourceBase) Resources() Resources {
	return r.Children
}

// ResourceSelector さくらのクラウド上で対象リソースを特定するための情報を提供する
type ResourceSelector struct {
	ID    types.ID `yaml:"id"`
	Names []string `yaml:"names"`
	Zones []string `yaml:"zone"` // グローバルリソースの場合はis1aまたは空とする TODO 要検討
}

func (rs *ResourceSelector) String() string {
	if rs != nil {
		return fmt.Sprintf("ID: %s, Names: %s, Zones: %s", rs.ID, rs.Names, rs.Zones)
	}
	return ""
}

func (rs *ResourceSelector) FindCondition() *sacloud.FindCondition {
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

type Desired interface {
	// Instruction 現在のリソースの状態から算出されたハンドラーへの指示の種類
	Instruction() handler.ResourceInstructions
	// ToRequest ハンドラーに渡すパラメータ、InstructionがNOOPやDELETEの場合はnilを返す
	ToRequest() *handler.Resource
}
