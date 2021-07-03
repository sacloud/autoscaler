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
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ResourceDefServerGroup struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`

	Name string `yaml:"name" validate:"required"` // {{ .Name }}{{ .Number }}
	Zone string `yaml:"zone" validate:"required,zone"`

	MinSize int `yaml:"min_size" validate:"min=0,ltefield=MaxSize"`
	MaxSize int `yaml:"max_size" validate:"min=0,gtecsfield=MinSize"`

	Plans []*ServerGroupPlan `yaml:"plans"`

	Template      *ServerGroupInstanceTemplate `yaml:"template" validate:"required"`
	ShutdownForce bool                         `yaml:"shutdown_force"`

	parent ResourceDefinition
}

func (d *ResourceDefServerGroup) String() string {
	return fmt.Sprintf("Zone: %s, Name: %s", d.Zone, d.Name)
}

func (d *ResourceDefServerGroup) Parent() ResourceDefinition {
	return d.parent
}

func (d *ResourceDefServerGroup) SetParent(parent ResourceDefinition) {
	d.parent = parent
}

func (d *ResourceDefServerGroup) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	for _, p := range d.Plans {
		if !(d.MinSize <= p.Size && p.Size <= d.MaxSize) {
			errors = multierror.Append(errors, fmt.Errorf("plan: plan.size must be between min_size and max_size: size:%d", p.Size))
		}
	}
	errors = multierror.Append(errors, d.Template.Validate(ctx, apiClient, d)...)

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
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

func (d *ResourceDefServerGroup) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
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

	var resources Resources
	for i := range cloudResources {
		instance := &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType: ResourceTypeServerGroupInstance,
			},
			apiClient:   apiClient,
			server:      cloudResources[i],
			zone:        d.Zone,
			def:         d,
			instruction: handler.ResourceInstructions_NOOP,
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
				resourceType: ResourceTypeServerGroupInstance,
			},
			apiClient: apiClient,
			server: &sacloud.Server{
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
		})
	}
	return resources, nil
}

func (d *ResourceDefServerGroup) desiredPlan(ctx *RequestContext, currentCount int) (*ServerGroupPlan, error) {
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
	return &ServerGroupPlan{Size: currentCount + 1}, nil
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
	return fmt.Sprintf(nameFormat, d.Name, index+1)
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

func (d *ResourceDefServerGroup) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.Server, error) {
	serverOp := sacloud.NewServerOp(apiClient)
	selector := &ResourceSelector{Names: []string{d.Name}}
	found, err := serverOp.Find(ctx, d.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}

	// 名前の昇順にソート
	sort.Slice(found.Servers, func(i, j int) bool {
		return found.Servers[i].Name < found.Servers[j].Name
	})
	return found.Servers, nil
}
