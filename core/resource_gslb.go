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

type ResourceGSLB struct {
	*ResourceBase

	apiClient sacloud.APICaller
	gslb      *sacloud.GSLB
	def       *ResourceDefGSLB
}

func NewResourceGSLB(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefGSLB, gslb *sacloud.GSLB) (*ResourceGSLB, error) {
	resource := &ResourceGSLB{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeGSLB},
		apiClient:    apiClient,
		gslb:         gslb,
		def:          def,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceGSLB) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	computed := &computedGSLB{
		instruction: handler.ResourceInstructions_NOOP,
		gslb:        &sacloud.GSLB{},
		resource:    r,
	}
	if err := mapconvDecoder.ConvertTo(r.gslb, computed.gslb); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	return computed, nil
}

func (r *ResourceGSLB) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.gslb.Tags, r.gslb.ID)
	if changed {
		gslbOp := sacloud.NewGSLBOp(r.apiClient)
		updated, err := gslbOp.Update(ctx, r.gslb.ID, &sacloud.GSLBUpdateRequest{
			Name:               r.gslb.Name,
			Description:        r.gslb.Description,
			Tags:               tags,
			IconID:             r.gslb.IconID,
			HealthCheck:        r.gslb.HealthCheck,
			DelayLoop:          r.gslb.DelayLoop,
			Weighted:           r.gslb.Weighted,
			SorryServer:        r.gslb.SorryServer,
			DestinationServers: r.gslb.DestinationServers,
			SettingsHash:       r.gslb.SettingsHash,
		})
		if err != nil {
			return err
		}
		r.gslb = updated
	}
	return nil
}

func (r *ResourceGSLB) refresh(ctx *RequestContext) error {
	gslbOp := sacloud.NewGSLBOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	gslb, err := gslbOp.Read(ctx, r.gslb.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := gslbOp.Find(ctx, FindConditionWithResourceIDTag(r.gslb.ID))
			if err != nil {
				return err
			}
			if len(found.GSLBs) == 0 {
				return fmt.Errorf("gslb not found with: Filter='%s'", resourceIDMarkerTag(r.gslb.ID))
			}
			if len(found.GSLBs) > 1 {
				return fmt.Errorf("invalid state: found multiple gslb with: Filter='%s'", resourceIDMarkerTag(r.gslb.ID))
			}
			gslb = found.GSLBs[0]
		} else {
			return err
		}
	}
	r.gslb = gslb
	return r.setResourceIDTag(ctx)
}
