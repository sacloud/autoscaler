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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ResourceTypes int

const (
	ResourceTypeUnknown ResourceTypes = iota
	ResourceTypeServer
	ResourceTypeServerGroup
	ResourceTypeEnhancedLoadBalancer
	ResourceTypeGSLB
	ResourceTypeDNS
)

// ResourceGroup 一意な名前をキーとするリソースのリスト
type ResourceGroup map[string]Resources

// Resources リソースのリスト
type Resources []Resource

// Resource Coreが扱うさくらのクラウド上のリソースを表す
//
// Core起動時のコンフィギュレーションから形成される
type Resource interface {
	Type() ResourceTypes
	Selector() *ResourceSelector
	Current() CurrentResource
	Desired() Desired
}

// ResourceSelector さくらのクラウド上で対象リソースを特定するための情報を提供する
type ResourceSelector struct {
	ID    types.ID
	Names []string
	Tags  []string
	Zone  string // グローバルリソースの場合はsacloud.APIDefaultZoneが入る // TODO 要検討
}

// CurrentResource リソースの現在の状態を示す
type CurrentResource interface {
	// Status 現在のリソースの状態
	Status() handler.ResourceStatus
	// Raw さくらのクラウドAPIからの戻り値(libsacloud v2のsacloud APIsの戻り値)
	Raw() interface{}
}

type Desired interface {
	// Raw CurrentResource.Raw()に対し更新すべきデータ差分(Up/Downの算出結果やtemplateの値など)を適用したもの
	//
	// Handlerはこの値から一部の値を取り出して処理することで、自身が関心のある項目のみを更新する
	// このためここに設定した値はHandlerの組み合わせ次第では実リソースへ適用されないことがある点に注意
	Raw() interface{}
}
