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
	"errors"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerPlan struct {
	Core   int `yaml:"core"`   // コア数
	Memory int `yaml:"memory"` // メモリサイズ(GiB)
}

// DefaultServerPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
//
// TODO 要検討
var DefaultServerPlans = []ServerPlan{
	{Core: 1, Memory: 1},
	{Core: 2, Memory: 4},
	{Core: 4, Memory: 8},
}

type ServerScalingOption struct {
	ShutdownForce bool `yaml:"shutdown_force"`
}

type Server struct {
	*ResourceBase `yaml:",inline"`
	DedicatedCPU  bool                `yaml:"dedicated_cpu"`
	PrivateHostID types.ID            `yaml:"private_host_id"`
	Plans         []ServerPlan        `yaml:"plans"`
	Option        ServerScalingOption `yaml:"option"`

	parent Resource `yaml:"-"`
}

func (s *Server) Validate() error {
	selector := s.Selector()
	if selector == nil {
		return errors.New("selector: required")
	}
	if len(selector.Zones) == 0 {
		return errors.New("selector.Zones: least one value required")
	}
	return nil
}

func (s *Server) Compute(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	var allComputed []Computed
	serverOp := sacloud.NewServerOp(apiClient)
	selector := s.Selector()

	for _, zone := range selector.Zones {
		found, err := serverOp.Find(ctx, zone, selector.FindCondition())
		if err != nil {
			return nil, fmt.Errorf("computing server status failed: %s", err)
		}
		for _, server := range found.Servers {
			computed, err := newComputedServer(ctx, s, zone, server)
			if err != nil {
				return nil, err
			}
			allComputed = append(allComputed, computed)
		}
	}

	if len(allComputed) == 0 {
		return nil, fmt.Errorf("server not found with selector: %s", selector.String())
	}

	s.ComputedCache = allComputed
	return allComputed, nil
}

func (s *Server) Parent() Resource {
	return s.parent
}

func (s *Server) SetParent(parent Resource) {
	s.parent = parent
}

type computedServer struct {
	instruction handler.ResourceInstructions
	server      *sacloud.Server
	zone        string
	newCPU      int
	newMemory   int
	resource    *Server // 算出元のResourceへの参照
}

func newComputedServer(ctx *Context, resource *Server, zone string, server *sacloud.Server) (*computedServer, error) {
	computed := &computedServer{
		instruction: handler.ResourceInstructions_NOOP,
		server:      &sacloud.Server{},
		zone:        zone,
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(server, computed.server); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan := computed.desiredPlan(ctx, server, resource.Plans)

	if plan != nil {
		computed.newCPU = plan.Core
		computed.newMemory = plan.Memory
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (c *computedServer) desiredPlan(ctx *Context, current *sacloud.Server, plans []ServerPlan) *ServerPlan {
	var fn func(i int) *ServerPlan

	if len(plans) == 0 {
		plans = DefaultServerPlans
	}

	// TODO s.Plansの並べ替えを考慮する

	if ctx.Request().refresh {
		// リフレッシュ時はプラン変更しない
		return nil
	}

	switch ctx.Request().requestType {
	case requestTypeUp:
		fn = func(i int) *ServerPlan {
			if i < len(plans)-1 {
				return &ServerPlan{
					Core:   plans[i+1].Core,
					Memory: plans[i+1].Memory,
				}
			}
			return nil
		}
	case requestTypeDown:
		fn = func(i int) *ServerPlan {
			if i > 0 {
				return &ServerPlan{
					Core:   plans[i-1].Core,
					Memory: plans[i-1].Memory,
				}
			}
			return nil
		}
	default:
		return nil // 到達しないはず
	}

	for i, plan := range plans {
		if plan.Core == current.CPU && plan.Memory == current.GetMemoryGB() {
			return fn(i)
		}
	}
	return nil
}

func (c *computedServer) ID() string {
	if c.server != nil {
		return c.server.ID.String()
	}
	return ""
}

func (c *computedServer) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedServer) parents() []*handler.Parent {
	if c.resource.parent != nil {
		return computedToParents(c.resource.parent.Computed())
	}
	return nil
}

func (c *computedServer) Current() *handler.Resource {
	if c.server != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Id:              c.server.ID.String(),
					Zone:            c.zone,
					Core:            uint32(c.server.CPU),
					Memory:          uint32(c.server.GetMemoryGB()),
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					PrivateHostId:   c.server.PrivateHostID.String(),
					AssignedNetwork: c.assignedNetwork(),
					Parents:         c.parents(),
					Option: &handler.ServerScalingOption{
						ShutdownForce: c.resource.Option.ShutdownForce,
					},
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
					Zone:            c.zone,
					Core:            uint32(c.newCPU),
					Memory:          uint32(c.newMemory),
					DedicatedCpu:    c.server.ServerPlanCommitment.IsDedicatedCPU(),
					PrivateHostId:   c.server.PrivateHostID.String(),
					AssignedNetwork: c.assignedNetwork(),
					Parents:         c.parents(),
					Option: &handler.ServerScalingOption{
						ShutdownForce: c.resource.Option.ShutdownForce,
					},
				},
			},
		}
	}
	return nil
}

func (c *computedServer) assignedNetwork() []*handler.NetworkInfo {
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
