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

// DefaultServerPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
var DefaultServerPlans = ResourcePlans{
	&ServerPlan{Core: 2, Memory: 4},
	&ServerPlan{Core: 4, Memory: 8},
	&ServerPlan{Core: 4, Memory: 16},
	&ServerPlan{Core: 8, Memory: 16},
	&ServerPlan{Core: 10, Memory: 24},
	&ServerPlan{Core: 10, Memory: 32},
	&ServerPlan{Core: 10, Memory: 48},
}

type ResourceServer struct {
	*ResourceBase

	apiClient sacloud.APICaller
	server    *sacloud.Server
	def       *ResourceDefServer
	zone      string
}

func NewResourceServer(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefServer, zone string, server *sacloud.Server) (*ResourceServer, error) {
	return &ResourceServer{
		ResourceBase: &ResourceBase{
			resourceType: ResourceTypeServer,
		},
		apiClient: apiClient,
		zone:      zone,
		server:    server,
		def:       def,
	}, nil
}

func (r *ResourceServer) String() string {
	if r == nil || r.server == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, Zone: %s, ID: %s, Name: %s}", r.Type(), r.zone, r.server.ID, r.server.Name)
}

func (r *ResourceServer) Compute(ctx *RequestContext, parent Computed, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	computed := &computedServer{
		instruction:   handler.ResourceInstructions_NOOP,
		server:        &sacloud.Server{},
		zone:          r.zone,
		shutdownForce: r.def.ShutdownForce,
		parent:        parent,
	}
	if err := mapconvDecoder.ConvertTo(r.server, computed.server); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	if !refresh && ctx.Request().resourceName == r.def.Name() {
		plan, err := r.desiredPlan(ctx)
		if err != nil {
			return nil, fmt.Errorf("computing desired plan failed: %s", err)
		}

		if plan != nil {
			computed.newCPU = plan.Core
			computed.newMemory = plan.Memory
			computed.instruction = handler.ResourceInstructions_UPDATE
		}
	}
	return computed, nil
}

func (r *ResourceServer) desiredPlan(ctx *RequestContext) (*ServerPlan, error) {
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

func (r *ResourceServer) refresh(ctx *RequestContext) error {
	server, err := query.ReadServer(ctx, r.apiClient, r.zone, r.server.ID)
	if err != nil {
		return err
	}
	r.server = server
	return nil
}
