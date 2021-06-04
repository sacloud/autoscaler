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

type GSLB struct {
	*ResourceBase `yaml:",inline"`
}

func (d *GSLB) Validate() error {
	// TODO 実装
	return nil
}

func (d *GSLB) Compute(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	var allComputed []Computed
	gslbOp := sacloud.NewGSLBOp(apiClient)
	selector := d.Selector()

	found, err := gslbOp.Find(ctx, selector.FindCondition())
	if err != nil {
		return nil, fmt.Errorf("computing GSLB status failed: %s", err)
	}
	for _, gslb := range found.GSLBs {
		computed, err := newComputedGSLB(ctx, d, gslb)
		if err != nil {
			return nil, err
		}
		allComputed = append(allComputed, computed)
	}

	if len(allComputed) == 0 {
		return nil, fmt.Errorf("gslb resource not found with selector: %s", selector.String())
	}

	d.ComputedCache = allComputed
	return allComputed, nil
}

type computedGSLB struct {
	instruction handler.ResourceInstructions
	gslb        *sacloud.GSLB
	resource    *GSLB // 算出元のResourceへの参照
}

func newComputedGSLB(ctx *Context, resource *GSLB, gslb *sacloud.GSLB) (*computedGSLB, error) {
	computed := &computedGSLB{
		instruction: handler.ResourceInstructions_NOOP,
		gslb:        &sacloud.GSLB{},
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(gslb, computed.gslb); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	return computed, nil
}

func (cg *computedGSLB) Instruction() handler.ResourceInstructions {
	return cg.instruction
}

func (cg *computedGSLB) Current() *handler.Resource {
	if cg.gslb != nil {
		var servers []*handler.GSLBServer
		for _, s := range cg.gslb.DestinationServers {
			servers = append(servers, &handler.GSLBServer{
				IpAddress: s.IPAddress,
				Enabled:   s.Enabled.Bool(),
				Weight:    uint32(s.Weight.Int()),
			})
		}

		return &handler.Resource{
			Resource: &handler.Resource_Gslb{
				Gslb: &handler.GSLB{
					Id:      cg.gslb.ID.String(),
					Fqdn:    cg.gslb.FQDN,
					Servers: servers,
				},
			},
		}
	}
	return nil
}

func (cg *computedGSLB) Desired() *handler.Resource {
	// GSLBリソースは基本的に参照専用なため常にCurrentを返すのみ
	return cg.Current()
}
