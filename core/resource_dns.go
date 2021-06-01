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
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type DNS struct {
	*ResourceBase `yaml:",inline"`
}

func (d *DNS) Validate() error {
	// TODO 実装
	return nil
}

func (d *DNS) Compute(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
	if len(d.ComputedCache) != 0 {
		return d.ComputedCache, nil
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	var allComputed []Computed
	dnsOp := sacloud.NewDNSOp(apiClient)
	selector := d.Selector()

	found, err := dnsOp.Find(ctx, selector.FindCondition())
	if err != nil {
		return nil, fmt.Errorf("computing DNS status failed: %s", err)
	}
	for _, dns := range found.DNS {
		computed, err := newComputedDNS(ctx, d, dns)
		if err != nil {
			return nil, err
		}
		allComputed = append(allComputed, computed)
	}

	if len(allComputed) == 0 {
		return nil, fmt.Errorf("dns resource not found with selector: %s", selector.String())
	}

	d.ComputedCache = allComputed
	return allComputed, nil
}

type computedDNS struct {
	instruction handler.ResourceInstructions
	dns         *sacloud.DNS
	resource    *DNS // 算出元のResourceへの参照
}

func newComputedDNS(ctx *Context, resource *DNS, dns *sacloud.DNS) (*computedDNS, error) {
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

func (cd *computedDNS) Instruction() handler.ResourceInstructions {
	return cd.instruction
}

func (cd *computedDNS) Current() *handler.Resource {
	if cd.dns != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Dns{
				Dns: &handler.DNS{
					Id:         cd.dns.ID.String(),
					Zone:       cd.dns.DNSZone,
					DnsServers: cd.dns.DNSNameServers,
				},
			},
		}
	}
	return nil
}

func (cd *computedDNS) Desired() *handler.Resource {
	// DNSリソースは基本的に参照専用なため常にCurrentを返すのみ
	return cd.Current()
}
