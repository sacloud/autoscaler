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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
)

type computedServer struct {
	instruction      handler.ResourceInstructions
	setupGracePeriod int

	server        *iaas.Server
	zone          string
	newCPU        int
	newMemory     int
	parent        Computed // 親Resourceのcomputed
	shutdownForce bool
}

func (c *computedServer) ID() string {
	if c.server != nil {
		return c.server.ID.String()
	}
	return ""
}

func (c *computedServer) Name() string {
	if c.server != nil {
		return c.server.Name
	}
	return ""
}

func (c *computedServer) Type() ResourceTypes {
	return ResourceTypeServer
}

func (c *computedServer) Zone() string {
	return c.zone
}

func (c *computedServer) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedServer) SetupGracePeriod() int {
	return c.setupGracePeriod
}

func (c *computedServer) parents() *handler.Parent {
	return computedToParents(c.parent)
}

func (c *computedServer) Current() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Id:              c.server.ID.String(),
					Name:            c.server.Name,
					Zone:            c.zone,
					Core:            uint32(c.server.CPU),
					Memory:          uint32(c.server.GetMemoryGB()),
					Gpu:             uint32(c.server.GPU),
					CpuModel:        c.server.ServerPlanCPUModel,
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
					ShutdownForce:   c.shutdownForce,
				},
			},
		}
	}
	return nil
}

func (c *computedServer) Desired() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Id:              c.server.ID.String(),
					Name:            c.server.Name,
					Zone:            c.zone,
					Core:            uint32(c.newCPU),
					Memory:          uint32(c.newMemory),
					Gpu:             uint32(c.server.GPU),
					CpuModel:        c.server.ServerPlanCPUModel,
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
					ShutdownForce:   c.shutdownForce,
				},
			},
		}
	}
	return nil
}

func (c *computedServer) assignedNetwork() []*handler.NetworkInfo {
	var info []*handler.NetworkInfo
	for i, nic := range c.server.Interfaces {
		info = append(info, assignedNetwork(nic, i))
	}
	return info
}
