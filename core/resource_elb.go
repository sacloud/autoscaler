// Copyright 2021-2022 The sacloud/autoscaler Authors
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
	"github.com/sacloud/libsacloud/v2/helper/query"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceELB struct {
	*ResourceBase

	apiClient sacloud.APICaller
	elb       *sacloud.ProxyLB
	def       *ResourceDefELB
}

func NewResourceELB(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefELB, elb *sacloud.ProxyLB) (*ResourceELB, error) {
	return &ResourceELB{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeELB},
		apiClient:    apiClient,
		elb:          elb,
		def:          def,
	}, nil
}

func (r *ResourceELB) String() string {
	if r == nil || r.elb == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, ID: %s, Name: %s}", r.Type(), r.elb.ID, r.elb.Name)
}

func (r *ResourceELB) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}
	var parent Computed
	if r.parent != nil {
		pc, err := r.parent.Compute(ctx, false)
		if err != nil {
			return nil, err
		}
		parent = pc
	}

	computed := &computedELB{
		instruction: handler.ResourceInstructions_NOOP,
		elb:         &sacloud.ProxyLB{},
		resource:    r,
		parent:      parent,
	}
	if err := mapconvDecoder.ConvertTo(r.elb, computed.elb); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	if !refresh && ctx.Request().resourceName == r.def.Name() {
		plan, err := r.desiredPlan(ctx)
		if err != nil {
			return nil, fmt.Errorf("computing desired plan failed: %s", err)
		}

		if plan != nil {
			computed.newCPS = plan.CPS
			computed.instruction = handler.ResourceInstructions_UPDATE
		}
	}
	return computed, nil
}

func (r *ResourceELB) desiredPlan(ctx *RequestContext) (*ELBPlan, error) {
	plans := r.def.resourcePlans()
	plan, err := desiredPlan(ctx, r.elb, plans)
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

func (r *ResourceELB) refresh(ctx *RequestContext) error {
	elb, err := query.ReadProxyLB(ctx, r.apiClient, r.elb.ID)
	if err != nil {
		return err
	}
	r.elb = elb
	return nil
}
