// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
)

type computedDNS struct {
	instruction      handler.ResourceInstructions
	setupGracePeriod int
	dns              *iaas.DNS
}

func (c *computedDNS) ID() string {
	if c.dns != nil {
		return c.dns.ID.String()
	}
	return ""
}

func (c *computedDNS) Name() string {
	if c.dns != nil {
		return c.dns.Name
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

func (c *computedDNS) SetupGracePeriod() int {
	return c.setupGracePeriod
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
