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
	return &ResourceLoadBalancer{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeLoadBalancer},
		apiClient:    apiClient,
		lb:           lb,
		def:          def,
		zone:         zone,
	}, nil
}

func (r *ResourceLoadBalancer) String() string {
	if r == nil || r.lb == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, Zone: %s, ID: %s, Name: %s}", r.Type(), r.zone, r.lb.ID, r.lb.Name)
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

func (r *ResourceLoadBalancer) refresh(ctx *RequestContext) error {
	lb, err := sacloud.NewLoadBalancerOp(r.apiClient).Read(ctx, r.zone, r.lb.ID)
	if err != nil {
		return err
	}
	r.lb = lb
	return nil
}
