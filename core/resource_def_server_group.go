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

	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud/types"

	"github.com/sacloud/autoscaler/handler"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDefServerGroup struct {
	*ResourceDefBase `yaml:",inline"`

	Name string `yaml:"name_prefix"` // {{ .Name }}{{ .Number }}
	Zone string `yaml:"zone"`

	MinSize       int                          `yaml:"min_size"`
	MaxSize       int                          `yaml:"max_size"`
	Template      *ServerGroupInstanceTemplate `yaml:"template"`
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

	if d.Zone == "" {
		errors = multierror.Append(errors, fmt.Errorf("zone: required"))
	} else {
		exist := false
		for _, z := range sacloud.SakuraCloudZones {
			if z == d.Zone {
				exist = true
				break
			}
		}
		if !exist {
			errors = multierror.Append(errors, fmt.Errorf("zone: invalid zone: %s", d.Zone))
		}

		if d.Template == nil {
			errors = multierror.Append(errors, fmt.Errorf("template: required"))
		} else {
			errors = multierror.Append(errors, d.Template.Validate(ctx, apiClient)...)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefServerGroup) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	// 現在のリソースを取得
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	// Min/MaxとUp/Downを考慮してサーバ数を決定
	serverCount := len(cloudResources)
	if ctx.Request().requestType == requestTypeUp {
		serverCount++
	} else {
		serverCount--
	}
	if serverCount > d.MaxSize {
		serverCount = d.MaxSize
	}
	if serverCount < d.MinSize {
		serverCount = d.MinSize
	}

	var resources Resources
	for i := range cloudResources {
		instance := &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType: ResourceTypeServerGroupInstance,
			},
			apiClient:    apiClient,
			server:       cloudResources[i],
			zone:         d.Zone,
			def:          d,
			instruction:  handler.ResourceInstructions_NOOP,
			indexInGroup: i,
		}
		if i >= serverCount {
			instance.instruction = handler.ResourceInstructions_DELETE
		}
		resources = append(resources, instance)
	}
	for len(resources) < serverCount {
		commitment := types.Commitments.Standard
		if d.Template.Plan.DedicatedCPU {
			commitment = types.Commitments.DedicatedCPU
		}
		resources = append(resources, &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType: ResourceTypeServerGroupInstance,
			},
			apiClient: apiClient,
			server: &sacloud.Server{
				Name:                 fmt.Sprintf("%s-%03d", d.Name, len(resources)+1), // TODO フォーマット指定可能にすべきか?
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
			indexInGroup: len(resources),
		})
	}
	return resources, nil
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
