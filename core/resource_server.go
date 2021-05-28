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
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerPlan struct {
	Core   int // コア数
	Memory int // メモリサイズ(GiB)
}

// DefaultServerPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
//
// TODO 要検討
var DefaultServerPlans = []ServerPlan{
	{Core: 1, Memory: 1},
	{Core: 2, Memory: 4},
	{Core: 4, Memory: 8},
}

type Server struct {
	*ResourceBase `yaml:",inline"`
	DedicatedCPU  bool         `yaml:"dedicated_cpu"`
	PrivateHostID types.ID     `yaml:"private_host_id"`
	Zone          string       `yaml:"zone"`
	Plans         []ServerPlan `yaml:"plans"`
	Wrappers      Resources    `yaml:"wrappers"`
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

func (s *Server) Desired(ctx *Context, apiClient sacloud.APICaller) (Desired, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	serverOp := sacloud.NewServerOp(apiClient)
	selector := s.Selector()
	var server *sacloud.Server
	for _, zone := range selector.Zones {
		fc := selector.FindCondition()
		found, err := serverOp.Find(ctx, zone, fc)
		if err != nil {
			return nil, fmt.Errorf("computing server status failed: %s", err)
		}
		if len(found.Servers) > 0 {
			server = found.Servers[0]
		}
	}

	if server == nil {
		return nil, fmt.Errorf("server not found with selector: %s", selector.String())
	}

	desired := &desiredServer{
		instruction: handler.ResourceInstructions_NOOP,
		server:      &sacloud.Server{},
		zone:        s.Zone,
	}
	if err := mapconvDecoder.ConvertTo(server, desired.server); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan := s.desiredPlan(ctx, server)

	if plan != nil {
		desired.server.CPU = plan.Core
		desired.server.MemoryMB = plan.Memory * size.GiB
		desired.instruction = handler.ResourceInstructions_UPDATE
	}

	return desired, nil
}

func (s *Server) desiredPlan(ctx *Context, current *sacloud.Server) *ServerPlan {
	var fn func(i int) *ServerPlan

	plans := s.Plans
	if len(plans) == 0 {
		plans = DefaultServerPlans
	}

	// TODO s.Plansの並べ替えを考慮する

	switch ctx.Request().requestType {
	case requestTypeUp:
		fn = func(i int) *ServerPlan {
			if i < len(plans) {
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

type desiredServer struct {
	instruction handler.ResourceInstructions
	server      *sacloud.Server
	zone        string
}

func (d *desiredServer) Instruction() handler.ResourceInstructions {
	return d.instruction
}

func (d *desiredServer) ToRequest() *handler.Resource {
	if d.server != nil {
		var assignedNetwork []*handler.NetworkInfo
		for _, nic := range d.server.Interfaces {
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
			})
		}

		return &handler.Resource{
			Resource: &handler.Resource_Server{
				Server: &handler.Server{
					Instruction:     d.Instruction(),
					Id:              d.server.ID.String(),
					Zone:            d.zone,
					Core:            uint32(d.server.CPU),
					Memory:          uint32(d.server.GetMemoryGB()),
					DedicatedCpu:    d.server.ServerPlanCommitment.IsDedicatedCPU(),
					PrivateHostId:   d.server.PrivateHostID.String(),
					AssignedNetwork: assignedNetwork,
				},
			},
		}
	}
	return nil
}
