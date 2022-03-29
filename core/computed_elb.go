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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
)

type computedELB struct {
	instruction handler.ResourceInstructions
	elb         *iaas.ProxyLB
	newCPS      int
	parent      Computed // 親リソースのComputed
}

func (c *computedELB) ID() string {
	if c.elb != nil {
		return c.elb.ID.String()
	}
	return ""
}

func (c *computedELB) Name() string {
	if c.elb != nil {
		return c.elb.Name
	}
	return ""
}

func (c *computedELB) Type() ResourceTypes {
	return ResourceTypeELB
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
					Name:             c.elb.Name,
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
					Name:             c.elb.Name,
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
