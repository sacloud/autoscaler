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
)

type ResourceDefLoadBalancer struct {
	*ResourceDefBase `yaml:",inline"`
	Selector         *MultiZoneSelector `yaml:"selector"`
}

func (d *ResourceDefLoadBalancer) String() string {
	return d.Selector.String()
}

func (d *ResourceDefLoadBalancer) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	if err := d.Selector.Validate(); err != nil {
		errors = multierror.Append(errors, err)
	} else {
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
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefLoadBalancer) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, lb := range cloudResources {
		r, err := NewResourceLoadBalancer(ctx, apiClient, d, lb.zone, lb.LoadBalancer)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefLoadBalancer) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sakuraCloudLoadBalancer, error) {
	lbOp := sacloud.NewLoadBalancerOp(apiClient)
	selector := d.Selector
	var results []*sakuraCloudLoadBalancer

	for _, zone := range selector.Zones {
		found, err := lbOp.Find(ctx, zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, lb := range found.LoadBalancers {
			results = append(results, &sakuraCloudLoadBalancer{LoadBalancer: lb, zone: zone})
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	return results, nil
}

type sakuraCloudLoadBalancer struct {
	*sacloud.LoadBalancer
	zone string
}
