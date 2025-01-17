// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"time"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/iaas-api-go"
)

// ResourceDefinition Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
type ResourceDefinition interface {
	Type() ResourceTypes // リソースの型
	Name() string
	String() string
	Validate(ctx context.Context, apiClient iaas.APICaller) []error

	// Compute 現在/あるべき姿を算出する
	//
	// TypeとSelectorを元にさくらのクラウドAPIを用いて実リソースを検索、Resourceを作成して返す
	Compute(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error)

	// LastModifiedAt この定義が対象とするリソース(群)の最終更新を返す
	LastModifiedAt(ctx *RequestContext, apiClient iaas.APICaller) (time.Time, error)
}

// ResourceDefBase 全てのリソース定義が実装すべき基本プロパティ
//
// Resourceの実装に埋め込む場合、Compute()でComputedCacheを設定すること
type ResourceDefBase struct {
	TypeName string `yaml:"type" validate:"required,oneof=EnhancedLoadBalancer ELB Router Server ServerGroup"`
	DefName  string `yaml:"name"`

	// セットアップのための猶予時間(秒数)
	// Handleされた後、セットアップの完了を待つためにこの秒数分待つ
	// 0の場合はTypeごとに定められたデフォルト値が用いられる。
	// デフォルト値:
	//    - Server: 60
	//    - 上記以外: 0
	SetupGracePeriodSec int `yaml:"setup_grace_period" validate:"omitempty,min=0,max=600"`
}

func (r *ResourceDefBase) Type() ResourceTypes {
	switch r.TypeName {
	case ResourceTypeServer.String():
		return ResourceTypeServer
	case ResourceTypeServerGroup.String():
		return ResourceTypeServerGroup
	case ResourceTypeServerGroupInstance.String():
		return ResourceTypeServerGroupInstance
	case ResourceTypeELB.String(), "ELB":
		return ResourceTypeELB
	case ResourceTypeGSLB.String():
		return ResourceTypeGSLB
	case ResourceTypeDNS.String():
		return ResourceTypeDNS
	case ResourceTypeRouter.String():
		return ResourceTypeRouter
	case ResourceTypeLoadBalancer.String():
		return ResourceTypeLoadBalancer
	default:
		panic("invalid typename: " + r.TypeName)
	}
}

func (r *ResourceDefBase) Name() string {
	return r.DefName
}

func (r *ResourceDefBase) SetupGracePeriod() int {
	sec := r.SetupGracePeriodSec
	if sec == 0 {
		v, ok := defaults.SetupGracePeriods[r.Type().String()]
		if ok {
			sec = v
		}
	}
	return sec
}
