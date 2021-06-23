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
}

func (d *ResourceDefLoadBalancer) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	selector := d.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	} else {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		} else {
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

func (d *ResourceDefLoadBalancer) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, lb := range cloudResources {
		r, err := NewResourceLoadBalancer(ctx, apiClient, d, d.Selector().Zone, lb)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefLoadBalancer) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.LoadBalancer, error) {
	lbOp := sacloud.NewLoadBalancerOp(apiClient)
	selector := d.Selector()

	found, err := lbOp.Find(ctx, selector.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.LoadBalancers) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	return found.LoadBalancers, nil
}
