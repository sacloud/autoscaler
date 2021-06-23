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
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDefDNS struct {
	*ResourceDefBase `yaml:",inline"`
}

func (d *ResourceDefDNS) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	selector := d.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	} else {
		if selector.Zone != "" {
			errors = multierror.Append(fmt.Errorf("selector.Zone: can not be specified for this resource"))
		}

		if _, err := d.findCloudResources(ctx, apiClient); err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefDNS) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, dns := range cloudResources {
		r, err := NewResourceDNS(ctx, apiClient, d, dns)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefDNS) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.DNS, error) {
	dnsOp := sacloud.NewDNSOp(apiClient)
	selector := d.Selector()

	found, err := dnsOp.Find(ctx, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.DNS) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	return found.DNS, nil
}
