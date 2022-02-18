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

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// DefaultELBPlans 各リソースで定義しなかった場合に利用されるデフォルトのプラン一覧
var DefaultELBPlans = ResourcePlans{
	&ELBPlan{CPS: 100},
	&ELBPlan{CPS: 500},
	&ELBPlan{CPS: 1_000},
	&ELBPlan{CPS: 5_000},
	&ELBPlan{CPS: 10_000},
	&ELBPlan{CPS: 50_000},
	&ELBPlan{CPS: 100_000},
	&ELBPlan{CPS: 400_000},
}

type ResourceDefELB struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`
	Selector         *ResourceSelector `yaml:"selector" validate:"required"`
	Plans            []*ELBPlan        `yaml:"plans"`
}

func (d *ResourceDefELB) String() string {
	return d.Selector.String()
}

func (d *ResourceDefELB) resourcePlans() ResourcePlans {
	if len(d.Plans) == 0 {
		return DefaultELBPlans
	}
	var plans ResourcePlans
	for _, p := range d.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (d *ResourceDefELB) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	if errs := d.validatePlans(ctx, apiClient); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	_, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefELB) validatePlans(_ context.Context, _ sacloud.APICaller) []error {
	var errors []error
	names := map[string]struct{}{}

	if len(d.Plans) == 1 {
		errors = append(errors, fmt.Errorf("at least two plans must be specified"))
		return errors
	}

	for _, p := range d.Plans {
		if p.Name != "" {
			if _, ok := names[p.Name]; ok {
				errors = append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
			}
			names[p.Name] = struct{}{}
		}

		if p.CPS != types.ProxyLBPlans.CPS100.Int() &&
			p.CPS != types.ProxyLBPlans.CPS500.Int() &&
			p.CPS != types.ProxyLBPlans.CPS1000.Int() &&
			p.CPS != types.ProxyLBPlans.CPS5000.Int() &&
			p.CPS != types.ProxyLBPlans.CPS10000.Int() &&
			p.CPS != types.ProxyLBPlans.CPS50000.Int() &&
			p.CPS != types.ProxyLBPlans.CPS100000.Int() &&
			p.CPS != types.ProxyLBPlans.CPS400000.Int() {
			errors = append(errors, fmt.Errorf("plan{CPS:%d} not found", p.CPS))
		}
	}
	return errors
}

func (d *ResourceDefELB) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, elb := range cloudResources {
		r, err := NewResourceELB(ctx, apiClient, d, elb)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefELB) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.ProxyLB, error) {
	elbOp := sacloud.NewProxyLBOp(apiClient)
	selector := d.Selector

	found, err := elbOp.Find(ctx, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.ProxyLBs) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	return found.ProxyLBs, nil
}
