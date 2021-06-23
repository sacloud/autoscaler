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
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ResourceDefServer struct {
	*ResourceDefBase `yaml:",inline"`
	DedicatedCPU     bool                `yaml:"dedicated_cpu"`
	Plans            []*ServerPlan       `yaml:"plans"`
	Option           ServerScalingOption `yaml:"option"`

	parent ResourceDefinition
}

type ServerScalingOption struct {
	ShutdownForce bool `yaml:"shutdown_force"`
}

func (d *ResourceDefServer) resourcePlans() ResourcePlans {
	if len(d.Plans) == 0 {
		return DefaultServerPlans
	}
	var plans ResourcePlans
	for _, p := range d.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (d *ResourceDefServer) Parent() ResourceDefinition {
	return d.parent
}

func (d *ResourceDefServer) SetParent(parent ResourceDefinition) {
	d.parent = parent
}

func (d *ResourceDefServer) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	selector := d.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	} else {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		} else {
			if errs := d.validatePlans(ctx, apiClient); len(errs) > 0 {
				errors = multierror.Append(errors, errs...)
			}

			resources, err := d.findCloudResources(ctx, apiClient)
			if err != nil {
				errors = multierror.Append(errors, err)
			}
			if len(d.children) > 0 && len(resources) > 1 {
				var names []string
				for _, r := range resources {
					names = append(names, fmt.Sprintf("{Zone:%s, ID:%s, Name:%s}", selector.Zone, r.ID, r.Name))
				}
				errors = multierror.Append(errors,
					fmt.Errorf("A resource definition with children must return one resource, but got multiple resources: definition: {Type:%s, Selector:%s}, got: %s",
						d.Type(), d.Selector(), fmt.Sprintf("[%s]", strings.Join(names, ",")),
					))
			}
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefServer) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(d.Plans) > 0 {
		if len(d.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		availablePlans, err := sacloud.NewServerPlanOp(apiClient).Find(ctx, d.Selector().Zone, nil)
		if err != nil {
			return []error{fmt.Errorf("validating server plan failed: %s", err)}
		}

		// for unique check: plan name
		names := map[string]struct{}{}

		errors := &multierror.Error{}
		for _, p := range d.Plans {
			if p.Name != "" {
				if _, ok := names[p.Name]; ok {
					errors = multierror.Append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
				}
				names[p.Name] = struct{}{}
			}

			exists := false
			for _, available := range availablePlans.ServerPlans {
				dedicatedCPU := available.Commitment == types.Commitments.DedicatedCPU
				if available.Availability.IsAvailable() && dedicatedCPU == d.DedicatedCPU &&
					available.CPU == p.Core && available.GetMemoryGB() == p.Memory {
					exists = true
					break
				}
			}
			if !exists {
				errors = multierror.Append(errors,
					fmt.Errorf("plan{zone:%s, core:%d, memory:%d, dedicated_cpu:%t} not exists", d.Selector().Zone, p.Core, p.Memory, d.DedicatedCPU))
			}
		}

		return errors.Errors
	}
	return nil
}

func (d *ResourceDefServer) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, server := range cloudResources {
		r, err := NewResourceServer(ctx, apiClient, d, d.Selector().Zone, server)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefServer) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.Server, error) {
	serverOp := sacloud.NewServerOp(apiClient)
	selector := d.Selector()

	// TODO セレクターに複数のゾーンを指定可能にしたらここも修正

	found, err := serverOp.Find(ctx, selector.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.Servers) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}

	return found.Servers, nil
}
