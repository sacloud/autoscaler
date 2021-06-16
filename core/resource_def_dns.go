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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDefDNS struct {
	*ResourceBase `yaml:",inline"`
}

func (d *ResourceDefDNS) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}
	selector := d.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	}
	if errors.Len() == 0 {
		if selector.Zone != "" {
			errors = multierror.Append(fmt.Errorf("selector.Zone: can not be specified for this resource"))
		}

		if _, err := d.findCloudResource(ctx, apiClient); err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (s *ResourceDefDNS) Fetch(ctx *RequestContext, apiClient sacloud.APICaller) ([]Resource, error) {
	// TODO Fetchを実装する
	return nil, nil
}

func (d *ResourceDefDNS) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error) {
	cloudResource, err := d.findCloudResource(ctx, apiClient)
	if err != nil {
		return nil, err
	}
	computed, err := newComputedDNS(ctx, d, cloudResource)
	if err != nil {
		return nil, err
	}

	d.ComputedCache = computed
	return computed, nil
}

func (d *ResourceDefDNS) findCloudResource(ctx context.Context, apiClient sacloud.APICaller) (*sacloud.DNS, error) {
	dnsOp := sacloud.NewDNSOp(apiClient)
	selector := d.Selector()

	found, err := dnsOp.Find(ctx, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing status failed: %s", err)
	}
	if len(found.DNS) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}
	if len(found.DNS) > 1 {
		return nil, fmt.Errorf("multiple resources found with selector: %s", selector.String())
	}
	return found.DNS[0], nil
}

type computedDNS struct {
	instruction handler.ResourceInstructions
	dns         *sacloud.DNS
	resource    *ResourceDefDNS // 算出元のResourceへの参照
}

func newComputedDNS(ctx *RequestContext, resource *ResourceDefDNS, dns *sacloud.DNS) (*computedDNS, error) {
	computed := &computedDNS{
		instruction: handler.ResourceInstructions_NOOP,
		dns:         &sacloud.DNS{},
		resource:    resource,
	}
	if err := mapconvDecoder.ConvertTo(dns, computed.dns); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	return computed, nil
}

func (c *computedDNS) ID() string {
	if c.dns != nil {
		return c.dns.ID.String()
	}
	return ""
}

func (c *computedDNS) Type() ResourceTypes {
	return ResourceTypeDNS
}

func (c *computedDNS) Zone() string {
	return ""
}

func (c *computedDNS) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedDNS) Current() *handler.Resource {
	if c.dns != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Dns{
				Dns: &handler.DNS{
					Id:         c.dns.ID.String(),
					Zone:       c.dns.DNSZone,
					DnsServers: c.dns.DNSNameServers,
				},
			},
		}
	}
	return nil
}

func (c *computedDNS) Desired() *handler.Resource {
	// DNSリソースは基本的に参照専用なため常にCurrentを返すのみ
	return c.Current()
}
