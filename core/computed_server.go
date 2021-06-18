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
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type computedServer2 struct {
	instruction handler.ResourceInstructions
	server      *sacloud.Server
	zone        string
	newCPU      int
	newMemory   int
	parent      Computed         // 親Resourceのcomputed
	resource    *ResourceServer2 // 算出元のResourceへの参照
}

func (c *computedServer2) ID() string {
	if c.server != nil {
		return c.server.ID.String()
	}
	return ""
}

func (c *computedServer2) Type() ResourceTypes {
	return ResourceTypeServer
}

func (c *computedServer2) Zone() string {
	return c.zone
}

func (c *computedServer2) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedServer2) parents() *handler.Parent {
	return computedToParents(c.parent)
}

func (c *computedServer2) Current() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Id:              c.server.ID.String(),
					Zone:            c.zone,
					Core:            uint32(c.server.CPU),
					Memory:          uint32(c.server.GetMemoryGB()),
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
					Option: &handler.ServerScalingOption{
						ShutdownForce: c.resource.def.Option.ShutdownForce,
					},
				},
			},
		}
	}
	return nil
}

func (c *computedServer2) Desired() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Id:              c.server.ID.String(),
					Zone:            c.zone,
					Core:            uint32(c.newCPU),
					Memory:          uint32(c.newMemory),
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
					Option: &handler.ServerScalingOption{
						ShutdownForce: c.resource.def.Option.ShutdownForce,
					},
				},
			},
		}
	}
	return nil
}

func (c *computedServer2) assignedNetwork() []*handler.NetworkInfo {
	var assignedNetwork []*handler.NetworkInfo
	for i, nic := range c.server.Interfaces {
		var ipAddress string
		if nic.SwitchScope == types.Scopes.Shared {
			ipAddress = nic.IPAddress
		} else {
			ipAddress = nic.UserIPAddress
		}
		assignedNetwork = append(assignedNetwork, &handler.NetworkInfo{
			IpAddress: ipAddress,
			Netmask:   uint32(nic.UserSubnetNetworkMaskLen),
			Gateway:   nic.UserSubnetDefaultRoute,
			Index:     uint32(i),
		})
	}
	return assignedNetwork
}
