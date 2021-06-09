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

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// DefaultRouterPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
//
// TODO 要検討
var DefaultRouterPlans = ResourcePlans{
	&RouterPlan{BandWidth: 100},
	&RouterPlan{BandWidth: 250},
	&RouterPlan{BandWidth: 500},
	&RouterPlan{BandWidth: 1000},
	&RouterPlan{BandWidth: 1500},
	&RouterPlan{BandWidth: 2000},
	&RouterPlan{BandWidth: 2500},
	&RouterPlan{BandWidth: 3000},
	&RouterPlan{BandWidth: 3500},
	&RouterPlan{BandWidth: 4000},
	&RouterPlan{BandWidth: 4500},
	&RouterPlan{BandWidth: 5000},
}

type Router struct {
	*ResourceBase `yaml:",inline"`
	Plans         []*RouterPlan `yaml:"plans"`
}

func (r *Router) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for _, p := range r.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (r *Router) Validate() error {
	// TODO 実装
	return nil
}

func (r *Router) Compute(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	routerOp := sacloud.NewInternetOp(apiClient)
	selector := r.Selector()

	found, err := routerOp.Find(ctx, selector.Zone, selector.FindCondition())
	if err != nil {
		return nil, fmt.Errorf("computing state failed: %s", err)
	}
	if len(found.Internet) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	if len(found.Internet) > 1 {
		return nil, fmt.Errorf("multiple resources found with selector: %s", selector.String())
	}

	computed, err := newComputedRouter(ctx, r, selector.Zone, found.Internet[0])
	if err != nil {
		return nil, err
	}

	r.ComputedCache = computed
	return computed, nil
}

type computedRouter struct {
	instruction  handler.ResourceInstructions
	router       *sacloud.Internet
	resource     *Router // 算出元のResourceへの参照
	zone         string
	newBandWidth int
}

func newComputedRouter(ctx *Context, resource *Router, zone string, router *sacloud.Internet) (*computedRouter, error) {
	computed := &computedRouter{
		instruction: handler.ResourceInstructions_NOOP,
		router:      &sacloud.Internet{},
		zone:        zone,
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(router, computed.router); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan, err := computed.desiredPlan(ctx, router, resource.resourcePlans())
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newBandWidth = plan.BandWidth
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (c *computedRouter) ID() string {
	if c.router != nil {
		return c.router.ID.String()
	}
	return ""
}

func (c *computedRouter) Type() ResourceTypes {
	return ResourceTypeRouter
}

func (c *computedRouter) Zone() string {
	return c.zone
}

func (c *computedRouter) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedRouter) Current() *handler.Resource {
	if c.router != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Router{
				Router: &handler.Router{
					Id:        c.router.ID.String(),
					Zone:      c.zone,
					BandWidth: uint32(c.router.BandWidthMbps),
				},
			},
		}
	}
	return nil
}

func (c *computedRouter) Desired() *handler.Resource {
	if c.router != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Router{
				Router: &handler.Router{
					Id:        c.router.ID.String(),
					Zone:      c.zone,
					BandWidth: uint32(c.newBandWidth),
				},
			},
		}
	}
	return nil
}

func (c *computedRouter) desiredPlan(ctx *Context, current *sacloud.Internet, plans ResourcePlans) (*RouterPlan, error) {
	if len(plans) == 0 {
		plans = DefaultRouterPlans
	}
	plan, err := desiredPlan(ctx, current, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*RouterPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return nil, nil
}
