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

type computedGSLB struct {
	instruction      handler.ResourceInstructions
	setupGracePeriod int

	gslb *iaas.GSLB
}

func (c *computedGSLB) ID() string {
	if c.gslb != nil {
		return c.gslb.ID.String()
	}
	return ""
}

func (c *computedGSLB) Name() string {
	if c.gslb != nil {
		return c.gslb.Name
	}
	return ""
}

func (c *computedGSLB) Type() ResourceTypes {
	return ResourceTypeGSLB
}

func (c *computedGSLB) Zone() string {
	return ""
}

func (c *computedGSLB) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedGSLB) SetupGracePeriod() int {
	return c.setupGracePeriod
}

func (c *computedGSLB) Current() *handler.Resource {
	if c.gslb != nil {
		var servers []*handler.GSLBServer
		for _, s := range c.gslb.DestinationServers {
			servers = append(servers, &handler.GSLBServer{
				IpAddress: s.IPAddress,
				Enabled:   s.Enabled.Bool(),
				Weight:    uint32(s.Weight.Int()),
			})
		}
		return &handler.Resource{
			Resource: &handler.Resource_Gslb{
				Gslb: &handler.GSLB{
					Id:      c.gslb.ID.String(),
					Name:    c.gslb.Name,
					Fqdn:    c.gslb.FQDN,
					Servers: servers,
				},
			},
		}
	}
	return nil
}

func (c *computedGSLB) Desired() *handler.Resource {
	// GSLBリソースは基本的に参照専用なため常にCurrentを返すのみ
	return c.Current()
}
