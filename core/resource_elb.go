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

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ELBPlan struct {
	// Name プラン名、リクエストからDesiredStateNameパラメータで指定される
	Name string `yaml:"name"`
	CPS  int    `yaml:"cps"`
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
	if selector.Zone != "" {
		return errors.New("selector.Zone: can not be specified for this resource")
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

func (e *EnhancedLoadBalancer) Compute(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}

	elbOp := sacloud.NewProxyLBOp(apiClient)
	selector := e.Selector()

	found, err := elbOp.Find(ctx, selector.FindCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.ProxyLBs) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	if len(found.ProxyLBs) > 1 {
		return nil, fmt.Errorf("multiple resources found with selector: %s", selector.String())
	}

	computed, err := newComputedELB(ctx, e, found.ProxyLBs[0])
	if err != nil {
		return nil, err
	}

	e.ComputedCache = computed
	return computed, nil
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

	plan, err := computed.desiredPlan(ctx, elb, resource.Plans)
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newCPS = plan.CPS
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (c *computedELB) desiredPlan(ctx *Context, current *sacloud.ProxyLB, plans []ELBPlan) (*ELBPlan, error) {
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].CPS < plans[j].CPS
	})

	req := ctx.Request()

	if req.refresh {
		// リフレッシュ時はプラン変更しない
		return nil, nil
	}

	// DesiredStateNameが指定されていたら該当プランを探す
	if req.desiredStateName != "" && req.desiredStateName != defaults.DesiredStateName {
		var found *ELBPlan
		for _, plan := range plans {
			if plan.Name == req.desiredStateName {
				found = &plan
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("desired plan %q not found: request: %s", req.desiredStateName, req.String())
		}

		switch req.requestType {
		case requestTypeUp:
			// foundとcurrentが同じ場合はOK
			if found.CPS < current.Plan.Int() {
				// Upリクエストなのに指定の名前のプランの方が小さいためプラン変更しない
				return nil, fmt.Errorf("desired plan %q is smaller than current plan", req.desiredStateName)
			}
		case requestTypeDown:
			// foundとcurrentが同じ場合はOK
			if found.CPS > current.Plan.Int() {
				// Downリクエストなのに指定の名前のプランの方が大きいためプラン変更しない
				return nil, fmt.Errorf("desired plan %q is larger than current plan", req.desiredStateName)
			}
		default:
			return nil, nil // 到達しない
		}
		return found, nil
	}

	var fn func(i int) *ELBPlan
	switch req.requestType {
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
		return nil, nil // 到達しないはず
	}

	for i, plan := range plans {
		if plan.CPS == current.Plan.Int() {
			return fn(i), nil
		}
	}
	return nil, fmt.Errorf("desired plan not found: request: %s", req.String())
}

func (c *computedELB) ID() string {
	if c.elb != nil {
		return c.elb.ID.String()
	}
	return ""
}

func (c *computedELB) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedELB) parent() *handler.Parent {
	if c.resource.parent != nil {
		return computedToParents(c.resource.parent.Computed())
	}
	return nil
}

func (c *computedELB) Current() *handler.Resource {
	if c.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               c.elb.ID.String(),
					Region:           c.elb.Region.String(),
					Plan:             uint32(c.elb.Plan.Int()),
					VirtualIpAddress: c.elb.VirtualIPAddress,
					Fqdn:             c.elb.FQDN,
					Parent:           c.parent(),
				},
			},
		}
	}
	return nil
}

func (c *computedELB) Desired() *handler.Resource {
	if c.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               c.elb.ID.String(),
					Region:           c.elb.Region.String(),
					Plan:             uint32(c.newCPS),
					VirtualIpAddress: c.elb.VirtualIPAddress,
					Fqdn:             c.elb.FQDN,
					Parent:           c.parent(),
				},
			},
		}
	}
	return nil
}
