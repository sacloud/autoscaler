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
)

type ResourceDefGSLB struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`
	Selector         *ResourceSelector `yaml:"selector" validate:"required"`
}

func (d *ResourceDefGSLB) String() string {
	return d.Selector.String()
}

func (d *ResourceDefGSLB) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	resources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		errors = multierror.Append(errors, err)
	}
	if len(d.children) > 0 && len(resources) > 1 {
		var names []string
		for _, r := range resources {
			names = append(names, fmt.Sprintf("{ID:%s, Name:%s}", r.ID, r.Name))
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

func (d *ResourceDefGSLB) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, gslb := range cloudResources {
		r, err := NewResourceGSLB(ctx, apiClient, d, gslb)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefGSLB) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.GSLB, error) {
	gslbOp := sacloud.NewGSLBOp(apiClient)
	selector := d.Selector

	found, err := gslbOp.Find(ctx, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.GSLBs) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	return found.GSLBs, nil
}
