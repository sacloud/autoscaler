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
	return &ResourceGSLB{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeGSLB},
		apiClient:    apiClient,
		gslb:         gslb,
		def:          def,
	}, nil
}

func (r *ResourceGSLB) String() string {
	if r == nil || r.gslb == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, ID: %s, Name: %s}", r.Type(), r.gslb.ID, r.gslb.Name)
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

func (r *ResourceGSLB) refresh(ctx *RequestContext) error {
	gslb, err := sacloud.NewGSLBOp(r.apiClient).Read(ctx, r.gslb.ID)
	if err != nil {
		return err
	}
	r.gslb = gslb
	return nil
}
