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
	"errors"
	"fmt"
	"sort"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ELBPlan struct {
	CPS int `yaml:"cps"`
}

// DefaultELBPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
var DefaultELBPlans = []ELBPlan{
	{CPS: 100},
	{CPS: 500},
	{CPS: 1_000},
	{CPS: 5_000},
	{CPS: 10_000},
	{CPS: 50_000},
	{CPS: 100_000},
	{CPS: 400_000},
}

type EnhancedLoadBalancer struct {
	*ResourceBase `yaml:",inline"`
	Plans         []ELBPlan `yaml:"plans"`
	parent        Resource
}

func (e *EnhancedLoadBalancer) Validate() error {
	selector := e.Selector()
	if selector == nil {
		return errors.New("selector: required")
	}
	if len(selector.Zones) != 0 {
		return errors.New("selector.Zones: can not be specified for this resource")
	}
	return nil
}

// Parent ChildResourceインターフェースの実装
func (e *EnhancedLoadBalancer) Parent() Resource {
	return e.parent
}

// SetParent ChildResourceインターフェースの実装
func (e *EnhancedLoadBalancer) SetParent(parent Resource) {
	e.parent = parent
}

func (e *EnhancedLoadBalancer) Compute(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}

	var allComputed []Computed
	elbOp := sacloud.NewProxyLBOp(apiClient)
	selector := e.Selector()

	found, err := elbOp.Find(ctx, selector.FindCondition())
	if err != nil {
		return nil, fmt.Errorf("computing ELB status failed: %s", err)
	}
	for _, elb := range found.ProxyLBs {
		computed, err := newComputedELB(ctx, e, elb)
		if err != nil {
			return nil, err
		}
		allComputed = append(allComputed, computed)
	}

	if len(allComputed) == 0 {
		return nil, fmt.Errorf("server not found with selector: %s", selector.String())
	}

	e.ComputedCache = allComputed
	return allComputed, nil
}

type computedELB struct {
	instruction handler.ResourceInstructions
	elb         *sacloud.ProxyLB
	newCPS      int
	resource    *EnhancedLoadBalancer // 算出元のResourceへの参照
}

func newComputedELB(ctx *Context, resource *EnhancedLoadBalancer, elb *sacloud.ProxyLB) (*computedELB, error) {
	computed := &computedELB{
		instruction: handler.ResourceInstructions_NOOP,
		elb:         &sacloud.ProxyLB{},
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(elb, computed.elb); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan := computed.desiredPlan(ctx, elb, resource.Plans)

	if plan != nil {
		computed.newCPS = plan.CPS
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (cs *computedELB) desiredPlan(ctx *Context, current *sacloud.ProxyLB, plans []ELBPlan) *ELBPlan {
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].CPS < plans[j].CPS
	})

	if ctx.Request().refresh {
		// リフレッシュ時はプラン変更しない
		return nil
	}

	var fn func(i int) *ELBPlan
	switch ctx.Request().requestType {
	case requestTypeUp:
		fn = func(i int) *ELBPlan {
			if i < len(plans)-1 {
				return &ELBPlan{
					CPS: plans[i+1].CPS,
				}
			}
			return nil
		}
	case requestTypeDown:
		fn = func(i int) *ELBPlan {
			if i > 0 {
				return &ELBPlan{
					CPS: plans[i-1].CPS,
				}
			}
			return nil
		}
	default:
		return nil // 到達しないはず
	}

	for i, plan := range plans {
		if plan.CPS == current.Plan.Int() {
			return fn(i)
		}
	}
	return nil
}

func (cs *computedELB) Instruction() handler.ResourceInstructions {
	return cs.instruction
}

func (cs *computedELB) parents() []*handler.Parent {
	if cs.resource.parent != nil {
		return computedToParents(cs.resource.parent.Computed())
	}
	return nil
}

func (cs *computedELB) Current() *handler.Resource {
	if cs.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               cs.elb.ID.String(),
					Region:           cs.elb.Region.String(),
					Plan:             uint32(cs.elb.Plan.Int()),
					VirtualIpAddress: cs.elb.VirtualIPAddress,
					Fqdn:             cs.elb.FQDN,
					Parents:          cs.parents(),
				},
			},
		}
	}
	return nil
}

func (cs *computedELB) Desired() *handler.Resource {
	if cs.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               cs.elb.ID.String(),
					Region:           cs.elb.Region.String(),
					Plan:             uint32(cs.newCPS),
					VirtualIpAddress: cs.elb.VirtualIPAddress,
					Fqdn:             cs.elb.FQDN,
					Parents:          cs.parents(),
				},
			},
		}
	}
	return nil
}
