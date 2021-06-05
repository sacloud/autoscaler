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

type RouterPlan struct {
	BandWidth int `yaml:"band_width"` // メモリサイズ(GiB)
}

// DefaultRouterPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
//
// TODO 要検討
var DefaultRouterPlans = []RouterPlan{
	{BandWidth: 100},
	{BandWidth: 250},
	{BandWidth: 500},
	{BandWidth: 1000},
	{BandWidth: 1500},
	{BandWidth: 2000},
	{BandWidth: 2500},
	{BandWidth: 3000},
	{BandWidth: 3500},
	{BandWidth: 4000},
	{BandWidth: 4500},
	{BandWidth: 5000},
}

type Router struct {
	*ResourceBase `yaml:",inline"`
	Plans         []RouterPlan `yaml:"plans"`
}

func (d *Router) Validate() error {
	// TODO 実装
	return nil
}

func (d *Router) Compute(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	routerOp := sacloud.NewInternetOp(apiClient)
	selector := d.Selector()

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

	computed, err := newComputedRouter(ctx, d, selector.Zone, found.Internet[0])
	if err != nil {
		return nil, err
	}

	d.ComputedCache = computed
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

	plan := computed.desiredPlan(ctx, router, resource.Plans)

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

func (c *computedRouter) desiredPlan(ctx *Context, current *sacloud.Internet, plans []RouterPlan) *RouterPlan {
	var fn func(i int) *RouterPlan

	if len(plans) == 0 {
		plans = DefaultRouterPlans
	}

	// TODO s.Plansの並べ替えを考慮する

	if ctx.Request().refresh {
		// リフレッシュ時はプラン変更しない
		return nil
	}

	switch ctx.Request().requestType {
	case requestTypeUp:
		fn = func(i int) *RouterPlan {
			if i < len(plans)-1 {
				return &RouterPlan{
					BandWidth: plans[i+1].BandWidth,
				}
			}
			return nil
		}
	case requestTypeDown:
		fn = func(i int) *RouterPlan {
			if i > 0 {
				return &RouterPlan{
					BandWidth: plans[i-1].BandWidth,
				}
			}
			return nil
		}
	default:
		return nil // 到達しないはず
	}

	for i, plan := range plans {
		if plan.BandWidth == current.BandWidthMbps {
			return fn(i)
		}
	}
	return nil
}
