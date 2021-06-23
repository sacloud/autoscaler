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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type computedELB struct {
	instruction handler.ResourceInstructions
	elb         *sacloud.ProxyLB
	newCPS      int
	parent      Computed      // 親リソースのComputed
	resource    *ResourceELB2 // 算出元のResourceへの参照
}

func (c *computedELB) ID() string {
	if c.elb != nil {
		return c.elb.ID.String()
	}
	return ""
}

func (c *computedELB) Type() ResourceTypes {
	return ResourceTypeEnhancedLoadBalancer
}

func (c *computedELB) Zone() string {
	return ""
}

func (c *computedELB) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedELB) parents() *handler.Parent {
	return computedToParents(c.parent)
}

func (c *computedELB) Current() *handler.Resource {
	if c.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               c.elb.ID.String(),
					Region:           c.elb.Region.String(),
					Plan:             uint32(c.elb.Plan.Int()),
					VirtualIpAddress: c.elb.VirtualIPAddress,
					Fqdn:             c.elb.FQDN,
					Parent:           c.parents(),
				},
			},
		}
	}
	return nil
}

func (c *computedELB) Desired() *handler.Resource {
	if c.elb != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Elb{
				Elb: &handler.ELB{
					Id:               c.elb.ID.String(),
					Region:           c.elb.Region.String(),
					Plan:             uint32(c.newCPS),
					VirtualIpAddress: c.elb.VirtualIPAddress,
					Fqdn:             c.elb.FQDN,
					Parent:           c.parents(),
				},
			},
		}
	}
	return nil
}