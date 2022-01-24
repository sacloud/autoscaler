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
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type computedServerGroupInstance struct {
	instruction handler.ResourceInstructions
	server      *sacloud.Server
	zone        string

	disks             []*handler.ServerGroupInstance_Disk
	diskEditParameter *handler.ServerGroupInstance_EditParameter

	cloudConfig string

	networkInterfaces []*handler.ServerGroupInstance_NIC

	parent   Computed                     // 親Resourceのcomputed
	resource *ResourceServerGroupInstance // 算出元のResourceへの参照
}

func (c *computedServerGroupInstance) ID() string {
	if c.server != nil && !c.server.ID.IsEmpty() {
		return c.server.ID.String()
	}
	return ""
}

func (c *computedServerGroupInstance) Name() string {
	if c.server != nil {
		return c.server.Name
	}
	return ""
}

func (c *computedServerGroupInstance) Type() ResourceTypes {
	return ResourceTypeServerGroupInstance
}

func (c *computedServerGroupInstance) Zone() string {
	return c.zone
}

func (c *computedServerGroupInstance) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedServerGroupInstance) parents() *handler.Parent {
	return computedToParents(c.parent)
}

func (c *computedServerGroupInstance) Current() *handler.Resource {
	// Current()とDesired()は同じ値を返す
	return c.computeResource()
}

func (c *computedServerGroupInstance) Desired() *handler.Resource {
	// Current()とDesired()は同じ値を返す
	return c.computeResource()
}

func (c *computedServerGroupInstance) computeResource() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_ServerGroupInstance{
				ServerGroupInstance: &handler.ServerGroupInstance{
					Parent:            c.parents(),
					Id:                c.server.ID.String(),
					Zone:              c.zone,
					Core:              uint32(c.server.CPU),
					Memory:            uint32(c.server.GetMemoryGB()),
					DedicatedCpu:      c.server.ServerPlanCommitment.IsDedicatedCPU(),
					PrivateHostId:     c.server.PrivateHostID.String(),
					Disks:             c.disks,
					EditParameter:     c.diskEditParameter,
					CloudConfig:       c.cloudConfig,
					NetworkInterfaces: c.networkInterfaces,
					CdRomId:           c.server.CDROMID.String(),
					InterfaceDriver:   c.server.InterfaceDriver.String(),
					Name:              c.server.Name,
					Tags:              c.server.Tags,
					Description:       c.server.Description,
					IconId:            c.server.IconID.String(),
					ShutdownForce:     c.resource.def.ShutdownForce,
				},
			},
		}
	}
	return nil
}
