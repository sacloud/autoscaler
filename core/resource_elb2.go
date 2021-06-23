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

type ResourceELB2 struct {
	*ResourceBase2

	apiClient sacloud.APICaller
	elb       *sacloud.ProxyLB
	def       *ResourceDefELB
}

func NewResourceELB(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefELB, elb *sacloud.ProxyLB) (*ResourceELB2, error) {
	resource := &ResourceELB2{
		ResourceBase2: &ResourceBase2{resourceType: ResourceTypeEnhancedLoadBalancer},
		apiClient:     apiClient,
		elb:           elb,
		def:           def,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceELB2) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
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

	plan, err := r.desiredPlan(ctx)
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newCPS = plan.CPS
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (r *ResourceELB2) desiredPlan(ctx *RequestContext) (*ELBPlan, error) {
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

func (r *ResourceELB2) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.elb.Tags, r.elb.ID)
	if changed {
		elbOp := sacloud.NewProxyLBOp(r.apiClient)
		updated, err := elbOp.Update(ctx, r.elb.ID, &sacloud.ProxyLBUpdateRequest{
			HealthCheck:   r.elb.HealthCheck,
			SorryServer:   r.elb.SorryServer,
			BindPorts:     r.elb.BindPorts,
			Servers:       r.elb.Servers,
			Rules:         r.elb.Rules,
			LetsEncrypt:   r.elb.LetsEncrypt,
			StickySession: r.elb.StickySession,
			Timeout:       r.elb.Timeout,
			Gzip:          r.elb.Gzip,
			SettingsHash:  r.elb.SettingsHash,
			Name:          r.elb.Name,
			Description:   r.elb.Description,
			Tags:          tags,
			IconID:        r.elb.IconID,
		})
		if err != nil {
			return err
		}
		r.elb = updated
	}
	return nil
}

func (r *ResourceELB2) refresh(ctx *RequestContext) error {
	elbOp := sacloud.NewProxyLBOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	elb, err := elbOp.Read(ctx, r.elb.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := elbOp.Find(ctx, FindConditionWithResourceIDTag(r.elb.ID))
			if err != nil {
				return err
			}
			if len(found.ProxyLBs) == 0 {
				return fmt.Errorf("elb not found with: Filter='%s'", resourceIDMarkerTag(r.elb.ID))
			}
			if len(found.ProxyLBs) > 1 {
				return fmt.Errorf("invalid state: found multiple elb with: Filter='%s'", resourceIDMarkerTag(r.elb.ID))
			}
			elb = found.ProxyLBs[0]
		} else {
			return err
		}
	}
	r.elb = elb
	return r.setResourceIDTag(ctx)
}
