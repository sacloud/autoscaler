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

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// DefaultRouterPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
//
// 東京第2ゾーンでのみ利用可能なプランは定義されていないため、利用したい場合は各リソース定義内で個別に定義する
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

type ResourceDefRouter struct {
	*ResourceBase `yaml:",inline"`
	Plans         []*RouterPlan `yaml:"plans"`
}

func (r *ResourceDefRouter) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for _, p := range r.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (r *ResourceDefRouter) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	selector := r.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	}
	if errors.Len() == 0 {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		}
	}

	if errors.Len() == 0 {
		if errs := r.validatePlans(ctx, apiClient); len(errs) > 0 {
			errors = multierror.Append(errors, errs...)
		}

		if _, err := r.findCloudResource(ctx, apiClient); err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", r.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (r *ResourceDefRouter) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(r.Plans) > 0 {
		if len(r.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		availablePlans, err := sacloud.NewInternetPlanOp(apiClient).Find(ctx, r.Selector().Zone, nil)
		if err != nil {
			return []error{fmt.Errorf("validating router plan failed: %s", err)}
		}

		// for unique check: plan name
		names := map[string]struct{}{}

		errors := &multierror.Error{}
		for _, p := range r.Plans {
			if p.Name != "" {
				if _, ok := names[p.Name]; ok {
					errors = multierror.Append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
				}
				names[p.Name] = struct{}{}
			}

			exists := false
			for _, available := range availablePlans.InternetPlans {
				if available.Availability.IsAvailable() && available.BandWidthMbps == p.BandWidth {
					exists = true
					break
				}
			}
			if !exists {
				errors = multierror.Append(errors, fmt.Errorf("plan{band_width:%d} not exists", p.BandWidth))
			}
		}
		return errors.Errors
	}
	return nil
}

func (r *ResourceDefRouter) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error) {
	cloudResource, err := r.findCloudResource(ctx, apiClient)
	if err != nil {
		return nil, err
	}
	computed, err := newComputedRouter(ctx, r, r.Selector().Zone, cloudResource)
	if err != nil {
		return nil, err
	}

	r.ComputedCache = computed
	return computed, nil
}

func (r *ResourceDefRouter) findCloudResource(ctx context.Context, apiClient sacloud.APICaller) (*sacloud.Internet, error) {
	routerOp := sacloud.NewInternetOp(apiClient)
	selector := r.Selector()

	found, err := routerOp.Find(ctx, selector.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing state failed: %s", err)
	}
	if len(found.Internet) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	if len(found.Internet) > 1 {
		return nil, fmt.Errorf("multiple resources found with selector: %s", selector.String())
	}

	return found.Internet[0], nil
}

type computedRouter struct {
	instruction  handler.ResourceInstructions
	router       *sacloud.Internet
	resource     *ResourceDefRouter // 算出元のResourceへの参照
	zone         string
	newBandWidth int
}

func newComputedRouter(ctx *RequestContext, resource *ResourceDefRouter, zone string, router *sacloud.Internet) (*computedRouter, error) {
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

func (c *computedRouter) desiredPlan(ctx *RequestContext, current *sacloud.Internet, plans ResourcePlans) (*RouterPlan, error) {
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
