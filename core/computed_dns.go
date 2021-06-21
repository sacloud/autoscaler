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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type computedDNS2 struct {
	instruction handler.ResourceInstructions
	dns         *sacloud.DNS
	resource    *ResourceDNS2 // 算出元のResourceへの参照
}

func (c *computedDNS2) ID() string {
	if c.dns != nil {
		return c.dns.ID.String()
	}
	return ""
}

func (c *computedDNS2) Type() ResourceTypes {
	return ResourceTypeDNS
}

func (c *computedDNS2) Zone() string {
	return ""
}

func (c *computedDNS2) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedDNS2) Current() *handler.Resource {
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

func (c *computedDNS2) Desired() *handler.Resource {
	// DNSリソースは基本的に参照専用なため常にCurrentを返すのみ
	return c.Current()
}
