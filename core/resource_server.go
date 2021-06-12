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
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// DefaultServerPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
var DefaultServerPlans = ResourcePlans{
	&ServerPlan{Core: 2, Memory: 4},
	&ServerPlan{Core: 4, Memory: 8},
	&ServerPlan{Core: 4, Memory: 16},
	&ServerPlan{Core: 8, Memory: 16},
	&ServerPlan{Core: 10, Memory: 24},
	&ServerPlan{Core: 10, Memory: 32},
	&ServerPlan{Core: 10, Memory: 48},
}

type ServerScalingOption struct {
	ShutdownForce bool `yaml:"shutdown_force"`
}

type Server struct {
	*ResourceBase `yaml:",inline"`
	DedicatedCPU  bool                `yaml:"dedicated_cpu"`
	Plans         []*ServerPlan       `yaml:"plans"`
	Option        ServerScalingOption `yaml:"option"`

	parent Resource `yaml:"-"`
}

func (s *Server) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for _, p := range s.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (s *Server) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	selector := s.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	}
	if errors.Len() == 0 {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		}
	}

	if errors.Len() == 0 {
		if errs := s.validatePlans(ctx, apiClient); len(errs) > 0 {
			errors = multierror.Append(errors, errs...)
		}

		if _, err := s.findCloudResource(ctx, apiClient); err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", s.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (s *Server) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(s.Plans) > 0 {
		if len(s.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		availablePlans, err := sacloud.NewServerPlanOp(apiClient).Find(ctx, s.Selector().Zone, nil)
		if err != nil {
			return []error{fmt.Errorf("validating server plan failed: %s", err)}
		}

		// for unique check: plan name
		names := map[string]struct{}{}

		errors := &multierror.Error{}
		for _, p := range s.Plans {
			if p.Name != "" {
				if _, ok := names[p.Name]; ok {
					errors = multierror.Append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
				}
				names[p.Name] = struct{}{}
			}

			exists := false
			for _, available := range availablePlans.ServerPlans {
				dedicatedCPU := available.Commitment == types.Commitments.DedicatedCPU
				if available.Availability.IsAvailable() && dedicatedCPU == s.DedicatedCPU &&
					available.CPU == p.Core && available.GetMemoryGB() == p.Memory {
					exists = true
					break
				}
			}
			if !exists {
				errors = multierror.Append(errors,
					fmt.Errorf("plan{zone:%s, core:%d, memory:%d, dedicated_cpu:%t} not exists", s.Selector().Zone, p.Core, p.Memory, s.DedicatedCPU))
			}
		}

		return errors.Errors
	}
	return nil
}

func (s *Server) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error) {
	cloudResource, err := s.findCloudResource(ctx, apiClient)
	if err != nil {
		return nil, err
	}
	computed, err := newComputedServer(ctx, s, s.Selector().Zone, cloudResource)
	if err != nil {
		return nil, err
	}

	s.ComputedCache = computed
	return computed, nil
}

func (s *Server) findCloudResource(ctx context.Context, apiClient sacloud.APICaller) (*sacloud.Server, error) {
	serverOp := sacloud.NewServerOp(apiClient)
	selector := s.Selector()

	found, err := serverOp.Find(ctx, selector.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.Servers) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	if len(found.Servers) > 1 {
		return nil, fmt.Errorf("multiple resources found with selector: %s", selector.String())
	}

	return found.Servers[0], nil
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

func newComputedServer(ctx *RequestContext, resource *Server, zone string, server *sacloud.Server) (*computedServer, error) {
	computed := &computedServer{
		instruction: handler.ResourceInstructions_NOOP,
		server:      &sacloud.Server{},
		zone:        zone,
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(server, computed.server); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	plan, err := computed.desiredPlan(ctx, server, resource.resourcePlans())
	if err != nil {
		return nil, fmt.Errorf("computing desired plan failed: %s", err)
	}

	if plan != nil {
		computed.newCPU = plan.Core
		computed.newMemory = plan.Memory
		computed.instruction = handler.ResourceInstructions_UPDATE
	}
	return computed, nil
}

func (c *computedServer) desiredPlan(ctx *RequestContext, current *sacloud.Server, plans ResourcePlans) (*ServerPlan, error) {
	if len(plans) == 0 {
		plans = DefaultServerPlans
	}
	plan, err := desiredPlan(ctx, current, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*ServerPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return nil, nil
}

func (c *computedServer) ID() string {
	if c.server != nil {
		return c.server.ID.String()
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

func (c *computedServer) parents() *handler.Parent {
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
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
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
					AssignedNetwork: c.assignedNetwork(),
					Parent:          c.parents(),
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
