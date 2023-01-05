// Copyright 2021-2023 The sacloud/autoscaler Authors
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
	"github.com/sacloud/iaas-api-go"
)

type ParentResource struct {
	*ResourceBase

	apiClient iaas.APICaller
	resource  SakuraCloudResource
	def       *ParentResourceDef
	zone      string
}

func NewParentResource(ctx *RequestContext, apiClient iaas.APICaller, def *ParentResourceDef, resource SakuraCloudResource, zone string) (*ParentResource, error) {
	return &ParentResource{
		ResourceBase: &ResourceBase{resourceType: def.Type()},
		apiClient:    apiClient,
		resource:     resource,
		def:          def,
		zone:         zone,
	}, nil
}

func (r *ParentResource) String() string {
	if r == nil || r.resource == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, ID: %s, Name: %s}", r.Type(), r.resource.GetID(), r.resource.GetName())
}

func (r *ParentResource) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	var computed Computed

	switch r.def.Type() {
	case ResourceTypeELB:
		v := &iaas.ProxyLB{}
		if err := mapconvDecoder.ConvertTo(r.resource, v); err != nil {
			return nil, fmt.Errorf("computing desired state failed: %s", err)
		}
		computed = &computedELB{
			instruction: handler.ResourceInstructions_NOOP,
			elb:         v,
		}
	case ResourceTypeGSLB:
		v := &iaas.GSLB{}
		if err := mapconvDecoder.ConvertTo(r.resource, v); err != nil {
			return nil, fmt.Errorf("computing desired state failed: %s", err)
		}
		computed = &computedGSLB{
			instruction: handler.ResourceInstructions_NOOP,
			gslb:        v,
		}
	case ResourceTypeDNS:
		v := &iaas.DNS{}
		if err := mapconvDecoder.ConvertTo(r.resource, v); err != nil {
			return nil, fmt.Errorf("computing desired state failed: %s", err)
		}
		computed = &computedDNS{
			instruction: handler.ResourceInstructions_NOOP,
			dns:         v,
		}
	case ResourceTypeRouter:
		v := &iaas.Internet{}
		if err := mapconvDecoder.ConvertTo(r.resource, v); err != nil {
			return nil, fmt.Errorf("computing desired state failed: %s", err)
		}
		computed = &computedRouter{
			instruction: handler.ResourceInstructions_NOOP,
			router:      v,
			zone:        r.zone,
		}
	case ResourceTypeLoadBalancer:
		v := &iaas.LoadBalancer{}
		if err := mapconvDecoder.ConvertTo(r.resource, v); err != nil {
			return nil, fmt.Errorf("computing desired state failed: %s", err)
		}
		computed = &computedLoadBalancer{
			instruction: handler.ResourceInstructions_NOOP,
			lb:          v,
			zone:        r.zone,
		}
	default:
		panic("got unexpected type")
	}

	return computed, nil
}

func (r *ParentResource) refresh(ctx *RequestContext) error {
	var found SakuraCloudResource
	var err error

	switch r.def.Type() {
	case ResourceTypeELB:
		op := iaas.NewProxyLBOp(r.apiClient)
		found, err = op.Read(ctx, r.resource.GetID())
		if err != nil {
			return fmt.Errorf("computing status failed: %s", err)
		}
	case ResourceTypeGSLB:
		op := iaas.NewGSLBOp(r.apiClient)
		found, err = op.Read(ctx, r.resource.GetID())
		if err != nil {
			return fmt.Errorf("computing status failed: %s", err)
		}
	case ResourceTypeDNS:
		op := iaas.NewDNSOp(r.apiClient)
		found, err = op.Read(ctx, r.resource.GetID())
		if err != nil {
			return fmt.Errorf("computing status failed: %s", err)
		}
	case ResourceTypeRouter:
		op := iaas.NewInternetOp(r.apiClient)
		found, err = op.Read(ctx, r.zone, r.resource.GetID())
		if err != nil {
			return fmt.Errorf("computing status failed: %s", err)
		}
	case ResourceTypeLoadBalancer:
		op := iaas.NewLoadBalancerOp(r.apiClient)
		found, err = op.Read(ctx, r.zone, r.resource.GetID())
		if err != nil {
			return fmt.Errorf("computing status failed: %s", err)
		}
	default:
		panic("got unexpected type")
	}

	r.resource = found
	return nil
}
