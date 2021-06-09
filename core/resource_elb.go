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

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// DefaultELBPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
var DefaultELBPlans = ResourcePlans{
	&ELBPlan{CPS: 100},
	&ELBPlan{CPS: 500},
	&ELBPlan{CPS: 1_000},
	&ELBPlan{CPS: 5_000},
	&ELBPlan{CPS: 10_000},
	&ELBPlan{CPS: 50_000},
	&ELBPlan{CPS: 100_000},
	&ELBPlan{CPS: 400_000},
}

type EnhancedLoadBalancer struct {
	*ResourceBase `yaml:",inline"`
	Plans         []*ELBPlan `yaml:"plans"`
	parent        Resource
}

func (e *EnhancedLoadBalancer) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for _, p := range e.Plans {
		plans = append(plans, p)
	}
	return plans
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

	plan, err := computed.desiredPlan(ctx, elb, resource.resourcePlans())
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newCPS = plan.CPS
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (c *computedELB) desiredPlan(ctx *Context, current *sacloud.ProxyLB, plans ResourcePlans) (*ELBPlan, error) {
	if len(plans) == 0 {
		plans = DefaultELBPlans
	}
	plan, err := desiredPlan(ctx, current, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*ELBPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return nil, nil
}

func (c *computedELB) ID() string {
	if c.elb != nil {
		return c.elb.ID.String()
	}
	return ""
}

func (c *computedELB) Type() ResourceTypes {
	return ResourceTypeEnhancedLoadBalancer
}

func (c *computedELB) Zone() string {
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
