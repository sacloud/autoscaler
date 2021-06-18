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
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ResourceDefServer struct {
	*ResourceDefBase `yaml:",inline"`
	DedicatedCPU     bool                `yaml:"dedicated_cpu"`
	Plans            []*ServerPlan       `yaml:"plans"`
	Option           ServerScalingOption `yaml:"option"`
}

func (s *ResourceDefServer) resourcePlans() ResourcePlans {
	if len(s.Plans) == 0 {
		return DefaultServerPlans
	}
	var plans ResourcePlans
	for _, p := range s.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (s *ResourceDefServer) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	selector := s.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	} else {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		}
	}

	if errors.Len() == 0 {
		if errs := s.validatePlans(ctx, apiClient); len(errs) > 0 {
			errors = multierror.Append(errors, errs...)
		}

		if _, err := s.findCloudResources(ctx, apiClient); err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", s.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (s *ResourceDefServer) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(s.Plans) > 0 {
		if len(s.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		availablePlans, err := sacloud.NewServerPlanOp(apiClient).Find(ctx, s.Selector().Zone, nil)
		if err != nil {
			return []error{fmt.Errorf("validating server plan failed: %s", err)}
		}

		// for unique check: plan name
		names := map[string]struct{}{}

		errors := &multierror.Error{}
		for _, p := range s.Plans {
			if p.Name != "" {
				if _, ok := names[p.Name]; ok {
					errors = multierror.Append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
				}
				names[p.Name] = struct{}{}
			}

			exists := false
			for _, available := range availablePlans.ServerPlans {
				dedicatedCPU := available.Commitment == types.Commitments.DedicatedCPU
				if available.Availability.IsAvailable() && dedicatedCPU == s.DedicatedCPU &&
					available.CPU == p.Core && available.GetMemoryGB() == p.Memory {
					exists = true
					break
				}
			}
			if !exists {
				errors = multierror.Append(errors,
					fmt.Errorf("plan{zone:%s, core:%d, memory:%d, dedicated_cpu:%t} not exists", s.Selector().Zone, p.Core, p.Memory, s.DedicatedCPU))
			}
		}

		return errors.Errors
	}
	return nil
}

func (s *ResourceDefServer) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
	cloudResources, err := s.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources2
	for _, server := range cloudResources {
		r, err := NewResourceServer(ctx, apiClient, s, s.Selector().Zone, server)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (s *ResourceDefServer) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.Server, error) {
	serverOp := sacloud.NewServerOp(apiClient)
	selector := s.Selector()

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
