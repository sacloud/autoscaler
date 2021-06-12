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
	"sort"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handler"
)

// Computed 各リソースが算出した希望する状態(Desired State)を示すインターフェース
//
// 一部の例外を除き1つのComputedに対し1つのさくらのクラウド上のリソースが対応する
type Computed interface {
	// Type このComputedが表すさくらのクラウド上のリソースの種別
	Type() ResourceTypes
	// ID このComputedが表すさくらのクラウド上のリソースのID、まだ存在しないリソースの場合は空文字を返す
	ID() string
	// Zone このComputedが表すさくらのクラウド上のリソースが属するゾーン名、グローバルリソースの場合は空文字を返す
	Zone() string
	// Instruction 現在のリソースの状態から算出されたハンドラーへの指示の種類
	Instruction() handler.ResourceInstructions
	// Current ハンドラーに渡すパラメータ、現在の状態を示す 現在存在しないリソースの場合はnilを返す
	Current() *handler.Resource
	// Desired ハンドラーに渡すパラメータ、InstructionがNOOPやDELETEの場合はnilを返す
	Desired() *handler.Resource
}

// ResourcePlan 垂直スケールをサポートするリソースでのプランを表すインターフェース
type ResourcePlan interface {
	// PlanName プラン名、Configurationで任意のプランに任意の名前をつけるために利用する
	PlanName() string
	// Equals 指定のリソースが該当プランと同じ値を持っているか判定する
	Equals(resource interface{}) bool
	// LessThan 指定のリソースが該当プランより小さい値であるかを判定する(境界は含めない)
	LessThan(resource interface{}) bool
	// LessThanPlan 指定のプランが該当プランより小さい値であるかを判定する(境界は含めない)
	LessThanPlan(plans ResourcePlan) bool
}

// ResourcePlans 垂直スケールをサポートするリソースでのプラン一覧を表すインターフェース
type ResourcePlans []ResourcePlan

// Sort ResourcePlan.LessThanPlanを用いて昇順にソートする
func (p *ResourcePlans) Sort() {
	plans := *p
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].LessThanPlan(plans[j])
	})
	*p = plans
}

// Next スケールアップ後のプランを取得する
//
// 該当プランが存在しない場合はnilを返す
func (p *ResourcePlans) Next(resource interface{}) ResourcePlan {
	next := false
	for _, plan := range *p {
		if plan.Equals(resource) || plan.LessThan(resource) {
			next = true
			continue
		}
		if next {
			return plan
		}
	}
	return nil
}

// Prev スケールダウン後のプランを取得する
//
// 該当プランが存在しない場合はnilを返す
func (p *ResourcePlans) Prev(resource interface{}) ResourcePlan {
	plans := *p
	var prev ResourcePlan
	for i, plan := range plans {
		if i > 0 && (plan.Equals(resource) || !plan.LessThan(resource)) {
			prev = plans[i-1]
			break
		}
	}
	if prev != nil && prev.Equals(resource) {
		return nil
	}
	return prev
}

func desiredPlan(ctx *RequestContext, current interface{}, plans ResourcePlans) (ResourcePlan, error) {
	plans.Sort()

	req := ctx.Request()
	if req.refresh {
		// リフレッシュ時はプラン変更しない
		return nil, nil
	}

	// DesiredStateNameが指定されていたら該当プランを探す
	if req.desiredStateName != "" && req.desiredStateName != defaults.DesiredStateName {
		var found ResourcePlan
		for _, plan := range plans {
			if plan.PlanName() == req.desiredStateName {
				found = plan
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("desired plan %q not found: request: %s", req.desiredStateName, req.String())
		}

		switch req.requestType {
		case requestTypeUp:
			// foundとcurrentが同じ場合はOK
			if found.LessThan(current) {
				// Upリクエストなのに指定の名前のプランの方が小さいためプラン変更しない
				return nil, fmt.Errorf("desired plan %q is smaller than current plan", req.desiredStateName)
			}
		case requestTypeDown:
			// foundとcurrentが同じ場合はOK
			if !(found.Equals(current) || found.LessThan(current)) {
				// Downリクエストなのに指定の名前のプランの方が大きいためプラン変更しない
				return nil, fmt.Errorf("desired plan %q is larger than current plan", req.desiredStateName)
			}
		default:
			return nil, nil // 到達しない
		}
		return found, nil
	}

	var desired ResourcePlan
	switch req.requestType {
	case requestTypeUp:
		desired = plans.Next(current)
	case requestTypeDown:
		desired = plans.Prev(current)
	default:
		return nil, nil // 到達しないはず
	}

	return desired, nil // nilの場合もあり得る
}
