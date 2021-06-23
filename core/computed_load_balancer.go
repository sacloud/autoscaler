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

type computedLoadBalancer struct {
	instruction handler.ResourceInstructions
	lb          *sacloud.LoadBalancer
	zone        string
	resource    *ResourceLoadBalancer // 算出元のResourceへの参照
}

func (c *computedLoadBalancer) ID() string {
	if c.lb != nil {
		return c.lb.ID.String()
	}
	return ""
}

func (c *computedLoadBalancer) Type() ResourceTypes {
	return ResourceTypeLoadBalancer
}

func (c *computedLoadBalancer) Zone() string {
	return c.zone
}

func (c *computedLoadBalancer) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedLoadBalancer) Current() *handler.Resource {
	if c.lb != nil {
		var vip []*handler.LoadBalancerVIP
		for _, v := range c.lb.VirtualIPAddresses {
			var servers []*handler.LoadBalancerServer
			for _, s := range v.Servers {
				servers = append(servers, &handler.LoadBalancerServer{
					IpAddress: s.IPAddress,
					Enabled:   s.Enabled.Bool(),
				})
			}
			vip = append(vip, &handler.LoadBalancerVIP{
				IpAddress: v.VirtualIPAddress,
				Port:      uint32(v.Port.Int()),
				Servers:   servers,
			})
		}
		return &handler.Resource{
			Resource: &handler.Resource_LoadBalancer{
				LoadBalancer: &handler.LoadBalancer{
					Id:                 c.lb.ID.String(),
					Zone:               c.zone,
					VirtualIpAddresses: vip,
				},
			},
		}
	}
	return nil
}

func (c *computedLoadBalancer) Desired() *handler.Resource {
	// LoadBalancerリソースは基本的に参照専用なため常にCurrentを返すのみ
	return c.Current()
}
