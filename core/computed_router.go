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

type computedRouter2 struct {
	instruction  handler.ResourceInstructions
	router       *sacloud.Internet
	zone         string
	newBandWidth int
	resource     *ResourceRouter2 // 算出元のResourceへの参照
}

func (c *computedRouter2) ID() string {
	if c.router != nil {
		return c.router.ID.String()
	}
	return ""
}

func (c *computedRouter2) Type() ResourceTypes {
	return ResourceTypeRouter
}

func (c *computedRouter2) Zone() string {
	return c.zone
}

func (c *computedRouter2) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *computedRouter2) Current() *handler.Resource {
	if c.router != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Router{
				Router: &handler.Router{
					Id:        c.router.ID.String(),
					Zone:      c.zone,
					BandWidth: uint32(c.router.BandWidthMbps),
				},
			},
		}
	}
	return nil
}

func (c *computedRouter2) Desired() *handler.Resource {
	if c.router != nil {
		return &handler.Resource{
			Resource: &handler.Resource_Router{
				Router: &handler.Router{
					Id:        c.router.ID.String(),
					Zone:      c.zone,
					BandWidth: uint32(c.newBandWidth),
				},
			},
		}
	}
	return nil
}
