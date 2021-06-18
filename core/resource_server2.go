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

type ResourceServer2 struct {
	*ResourceBase2

	apiClient sacloud.APICaller
	server    *sacloud.Server
	def       *ResourceDefServer
	zone      string
}

func NewResourceServer(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefServer, zone string, server *sacloud.Server) (*ResourceServer2, error) {
	resource := &ResourceServer2{
		ResourceBase2: &ResourceBase2{resourceType: ResourceTypeServer},
		apiClient:     apiClient,
		zone:          zone,
		server:        server,
		def:           def,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceServer2) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
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

	computed := &computedServer2{
		instruction: handler.ResourceInstructions_NOOP,
		server:      &sacloud.Server{},
		zone:        r.zone,
		resource:    r,
		parent:      parent,
	}
	if err := mapconvDecoder.ConvertTo(r.server, computed.server); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan, err := r.desiredPlan(ctx)
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newCPU = plan.Core
		computed.newMemory = plan.Memory
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (r *ResourceServer2) desiredPlan(ctx *RequestContext) (*ServerPlan, error) {
	plans := r.def.resourcePlans()
	plan, err := desiredPlan(ctx, r.server, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*ServerPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return nil, nil
}

func (r *ResourceServer2) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.server.Tags, r.server.ID)
	if changed {
		serverOp := sacloud.NewServerOp(r.apiClient)
		updated, err := serverOp.Update(ctx, r.zone, r.server.ID, &sacloud.ServerUpdateRequest{
			Name:            r.server.Name,
			Description:     r.server.Description,
			Tags:            tags,
			IconID:          r.server.IconID,
			PrivateHostID:   r.server.PrivateHostID,
			InterfaceDriver: r.server.InterfaceDriver,
		})
		if err != nil {
			return err
		}
		r.server = updated
	}
	return nil
}

func (r *ResourceServer2) refresh(ctx *RequestContext) error {
	serverOp := sacloud.NewServerOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	server, err := serverOp.Read(ctx, r.zone, r.server.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := serverOp.Find(ctx, r.zone, FindConditionWithResourceIDTag(r.server.ID))
			if err != nil {
				return err
			}
			if len(found.Servers) == 0 {
				return fmt.Errorf("server not found with: Filter='%s'", resourceIDMarkerTag(r.server.ID))
			}
			if len(found.Servers) > 1 {
				return fmt.Errorf("invalid state: found multiple server with: Filter='%s'", resourceIDMarkerTag(r.server.ID))
			}
			server = found.Servers[0]
		} else {
			return err
		}
	}
	r.server = server
	return r.setResourceIDTag(ctx)
}
