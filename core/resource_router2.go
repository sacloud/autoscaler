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

type ResourceRouter2 struct {
	*ResourceBase2

	apiClient sacloud.APICaller
	router    *sacloud.Internet
	def       *ResourceDefRouter
	zone      string
}

func NewResourceRouter(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefRouter, zone string, router *sacloud.Internet) (*ResourceRouter2, error) {
	resource := &ResourceRouter2{
		ResourceBase2: &ResourceBase2{resourceType: ResourceTypeRouter},
		apiClient:     apiClient,
		zone:          zone,
		router:        router,
		def:           def,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceRouter2) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	computed := &computedRouter2{
		instruction: handler.ResourceInstructions_NOOP,
		router:      &sacloud.Internet{},
		zone:        r.zone,
		resource:    r,
	}
	if err := mapconvDecoder.ConvertTo(r.router, computed.router); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan, err := r.desiredPlan(ctx)
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newBandWidth = plan.BandWidth
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (r *ResourceRouter2) desiredPlan(ctx *RequestContext) (*RouterPlan, error) {
	plans := r.def.resourcePlans()
	plan, err := desiredPlan(ctx, r.router, plans)
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

func (r *ResourceRouter2) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.router.Tags, r.router.ID)
	if changed {
		routerOp := sacloud.NewInternetOp(r.apiClient)
		updated, err := routerOp.Update(ctx, r.zone, r.router.ID, &sacloud.InternetUpdateRequest{
			Name:        r.router.Name,
			Description: r.router.Description,
			Tags:        tags,
			IconID:      r.router.IconID,
		})
		if err != nil {
			return err
		}
		r.router = updated
	}
	return nil
}

func (r *ResourceRouter2) refresh(ctx *RequestContext) error {
	routerOp := sacloud.NewInternetOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	router, err := routerOp.Read(ctx, r.zone, r.router.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := routerOp.Find(ctx, r.zone, FindConditionWithResourceIDTag(r.router.ID))
			if err != nil {
				return err
			}
			if len(found.Internet) == 0 {
				return fmt.Errorf("router not found with: Filter='%s'", resourceIDMarkerTag(r.router.ID))
			}
			if len(found.Internet) > 1 {
				return fmt.Errorf("invalid state: found multiple router with: Filter='%s'", resourceIDMarkerTag(r.router.ID))
			}
			router = found.Internet[0]
		} else {
			return err
		}
	}
	r.router = router
	return r.setResourceIDTag(ctx)
}
