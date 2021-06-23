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

type ResourceLoadBalancer struct {
	*ResourceBase

	apiClient sacloud.APICaller
	lb        *sacloud.LoadBalancer
	def       *ResourceDefLoadBalancer
	zone      string
}

func NewResourceLoadBalancer(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefLoadBalancer, zone string, lb *sacloud.LoadBalancer) (*ResourceLoadBalancer, error) {
	resource := &ResourceLoadBalancer{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeLoadBalancer},
		apiClient:    apiClient,
		lb:           lb,
		def:          def,
		zone:         zone,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceLoadBalancer) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	computed := &computedLoadBalancer{
		instruction: handler.ResourceInstructions_NOOP,
		lb:          &sacloud.LoadBalancer{},
		resource:    r,
		zone:        r.zone,
	}
	if err := mapconvDecoder.ConvertTo(r.lb, computed.lb); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	return computed, nil
}

func (r *ResourceLoadBalancer) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.lb.Tags, r.lb.ID)
	if changed {
		lbOp := sacloud.NewLoadBalancerOp(r.apiClient)
		updated, err := lbOp.Update(ctx, r.zone, r.lb.ID, &sacloud.LoadBalancerUpdateRequest{
			Name:               r.lb.Name,
			Description:        r.lb.Description,
			Tags:               tags,
			IconID:             r.lb.IconID,
			VirtualIPAddresses: r.lb.VirtualIPAddresses,
			SettingsHash:       r.lb.SettingsHash,
		})
		if err != nil {
			return err
		}
		r.lb = updated
	}
	return nil
}

func (r *ResourceLoadBalancer) refresh(ctx *RequestContext) error {
	lbOp := sacloud.NewLoadBalancerOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	lb, err := lbOp.Read(ctx, r.zone, r.lb.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := lbOp.Find(ctx, r.zone, FindConditionWithResourceIDTag(r.lb.ID))
			if err != nil {
				return err
			}
			if len(found.LoadBalancers) == 0 {
				return fmt.Errorf("lb not found with: Filter='%s'", resourceIDMarkerTag(r.lb.ID))
			}
			if len(found.LoadBalancers) > 1 {
				return fmt.Errorf("invalid state: found multiple lb with: Filter='%s'", resourceIDMarkerTag(r.lb.ID))
			}
			lb = found.LoadBalancers[0]
		} else {
			return err
		}
	}
	r.lb = lb
	return r.setResourceIDTag(ctx)
}
