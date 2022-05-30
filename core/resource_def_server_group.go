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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
)

type ResourceDefServerGroup struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`

	ServerNamePrefix string `yaml:"server_name_prefix"` // NameまたはServerNamePrefixが必須
	Zone             string `yaml:"zone" validate:"required,zone"`

	MinSize int `yaml:"min_size" validate:"min=0,ltefield=MaxSize"`
	MaxSize int `yaml:"max_size" validate:"min=0,gtecsfield=MinSize"`

	Plans []*ServerGroupPlan `yaml:"plans"`

	Template      *ServerGroupInstanceTemplate `yaml:"template" validate:"required"`
	ShutdownForce bool                         `yaml:"shutdown_force"`

	ParentDef *ParentResourceDef `yaml:"parent"`
}

func (d *ResourceDefServerGroup) String() string {
	return fmt.Sprintf("Zone: %s, Name: %s", d.Zone, d.Name())
}

func (d *ResourceDefServerGroup) namePrefix() string {
	if d.ServerNamePrefix != "" {
		return d.ServerNamePrefix
	}
	return d.Name()
}

func (d *ResourceDefServerGroup) Validate(ctx context.Context, apiClient iaas.APICaller) []error {
	errors := &multierror.Error{}
	if d.namePrefix() == "" {
		errors = multierror.Append(errors, fmt.Errorf("name or server_name_prefix: required"))
	}
	for _, p := range d.Plans {
		if !(d.MinSize <= p.Size && p.Size <= d.MaxSize) {
			errors = multierror.Append(errors, fmt.Errorf("plan: plan.size must be between min_size and max_size: size:%d", p.Size))
		}
	}
	errors = multierror.Append(errors, d.Template.Validate(ctx, apiClient, d)...)
	if d.ParentDef != nil {
		errors = multierror.Append(errors, d.ParentDef.Validate(ctx, apiClient, d.Zone)...)
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefServerGroup) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for size := d.MinSize; size <= d.MaxSize; size++ {
		plan := &ServerGroupPlan{Size: size}
		for _, p := range d.Plans {
			if p.Size == plan.Size {
				plan.Name = p.Name
				break
			}
		}
		plans = append(plans, plan)
	}
	return plans
}

func (d *ResourceDefServerGroup) Compute(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error) {
	ctx = ctx.WithZone(d.Zone)

	// 現在のリソースを取得
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	// Min/MaxとUp/Downを考慮してサーバ数を決定
	plan, err := d.desiredPlan(ctx, len(cloudResources))
	if err != nil {
		return nil, err
	}

	var parent Resource
	if d.ParentDef != nil {
		parents, err := d.ParentDef.Compute(ctx, apiClient)
		if err != nil {
			return nil, err
		}
		if len(parents) != 1 {
			return nil, fmt.Errorf("got invalid parent resources: %#+v", parents)
		}
		parent = parents[0]
	}

	var resources Resources
	for i := range cloudResources {
		instance := &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType:     ResourceTypeServerGroupInstance,
				setupGracePeriod: d.SetupGracePeriod(),
			},
			apiClient:   apiClient,
			server:      cloudResources[i],
			zone:        d.Zone,
			def:         d,
			instruction: handler.ResourceInstructions_NOOP,
			parent:      parent,
		}
		instance.indexInGroup = d.resourceIndex(instance)
		if i >= plan.Size {
			instance.instruction = handler.ResourceInstructions_DELETE
		}
		resources = append(resources, instance)
	}

	for len(resources) < plan.Size {
		commitment := types.Commitments.Standard
		if d.Template.Plan.DedicatedCPU {
			commitment = types.Commitments.DedicatedCPU
		}
		serverName, index := d.determineServerName(resources)
		resources = append(resources, &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType:     ResourceTypeServerGroupInstance,
				setupGracePeriod: d.SetupGracePeriod(),
			},
			apiClient: apiClient,
			server: &iaas.Server{
				Name:                 serverName,
				Tags:                 d.Template.Tags,
				Description:          d.Template.Description,
				IconID:               types.StringID(d.Template.IconID),
				CDROMID:              types.StringID(d.Template.CDROMID),
				PrivateHostID:        types.StringID(d.Template.PrivateHostID),
				InterfaceDriver:      d.Template.InterfaceDriver,
				CPU:                  d.Template.Plan.Core,
				MemoryMB:             d.Template.Plan.Memory * size.GiB,
				ServerPlanCommitment: commitment,
			},
			zone:         d.Zone,
			def:          d,
			instruction:  handler.ResourceInstructions_CREATE,
			indexInGroup: index,
			parent:       parent,
		})
	}
	return resources, nil
}

func (d *ResourceDefServerGroup) desiredPlan(ctx *RequestContext, currentCount int) (*ServerGroupPlan, error) {
	if ctx.Request().resourceName != d.Name() {
		return &ServerGroupPlan{Size: currentCount}, nil
	}
	plans := d.resourcePlans()
	plan, err := desiredPlan(ctx, currentCount, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*ServerGroupPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return &ServerGroupPlan{Size: currentCount}, nil
}

func (d *ResourceDefServerGroup) resourceIndex(resource Resource) int {
	for i := 0; i < d.MaxSize; i++ {
		name := d.serverNameByIndex(i)
		if resource.(*ResourceServerGroupInstance).server.Name == name {
			return i
		}
	}
	return 0
}

func (d *ResourceDefServerGroup) serverNameByIndex(index int) string {
	nameFormat := "%s-%03d" // TODO フォーマット指定可能にする
	return fmt.Sprintf(nameFormat, d.namePrefix(), index+1)
}

func (d *ResourceDefServerGroup) determineServerName(resources Resources) (string, int) {
	for i := range resources {
		name := d.serverNameByIndex(i)
		exist := false
		for _, r := range resources {
			if r.(*ResourceServerGroupInstance).server.Name == name {
				exist = true
				break
			}
		}
		if !exist {
			return name, i
		}
	}
	return d.serverNameByIndex(len(resources)), len(resources)
}

func (d *ResourceDefServerGroup) findCloudResources(ctx context.Context, apiClient iaas.APICaller) ([]*iaas.Server, error) {
	serverOp := iaas.NewServerOp(apiClient)
	selector := &ResourceSelector{Names: []string{d.namePrefix()}}
	found, err := serverOp.Find(ctx, d.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}

	// Nameとd.namePrefix()が前方一致するリソースだけに絞る
	servers := d.filterCloudServers(found.Servers)

	// 名前の昇順にソート
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})
	return servers, nil
}

func (d *ResourceDefServerGroup) filterCloudServers(servers []*iaas.Server) []*iaas.Server {
	var filtered []*iaas.Server
	for _, server := range servers {
		if strings.HasPrefix(server.Name, d.namePrefix()) {
			filtered = append(filtered, server)
		}
	}
	return filtered
}
