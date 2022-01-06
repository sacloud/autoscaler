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
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ResourceDefServer struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`
	Selector         *MultiZoneSelector `yaml:"selector" validate:"required"`

	DedicatedCPU  bool          `yaml:"dedicated_cpu"`
	Plans         []*ServerPlan `yaml:"plans"`
	ShutdownForce bool          `yaml:"shutdown_force"`

	parent ResourceDefinition
}

func (d *ResourceDefServer) String() string {
	return d.Selector.String()
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
			names = append(names, fmt.Sprintf("{Zone:%s, ID:%s, Name:%s}", r.zone, r.ID, r.Name))
		}
		errors = multierror.Append(errors,
			fmt.Errorf("A resource definition with children must return one resource, but got multiple resources: definition: {Type:%s, Selector:%s}, got: %s",
				d.Type(), d.Selector, fmt.Sprintf("[%s]", strings.Join(names, ",")),
			))
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefServer) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(d.Plans) > 0 {
		if len(d.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		errors := &multierror.Error{}
		for _, zone := range d.Selector.Zones {
			availablePlans, err := sacloud.NewServerPlanOp(apiClient).Find(ctx, zone, nil)
			if err != nil {
				return []error{fmt.Errorf("validating server plan failed: %s", err)}
			}

			// for unique check: plan name
			names := map[string]struct{}{}

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
						fmt.Errorf("plan{zone:%s, core:%d, memory:%d, dedicated_cpu:%t} not exists", zone, p.Core, p.Memory, d.DedicatedCPU))
				}
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
		r, err := NewResourceServer(ctx, apiClient, d, server.zone, server.Server)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefServer) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sakuraCloudServer, error) {
	serverOp := sacloud.NewServerOp(apiClient)
	selector := d.Selector
	var results []*sakuraCloudServer

	for _, zone := range selector.Zones {
		found, err := serverOp.Find(ctx, zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, s := range found.Servers {
			results = append(results, &sakuraCloudServer{Server: s, zone: zone})
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}

	return results, nil
}

type sakuraCloudServer struct {
	*sacloud.Server
	zone string
}
