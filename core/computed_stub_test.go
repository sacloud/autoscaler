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
)

type stubComputed struct {
	id               string
	name             string
	zone             string
	typ              ResourceTypes
	instruction      handler.ResourceInstructions
	setupGracePeriod int
	current          *handler.Resource
	desired          *handler.Resource
}

func (c *stubComputed) ID() string {
	return c.id
}

func (c *stubComputed) Name() string {
	return c.name
}

func (c *stubComputed) Type() ResourceTypes {
	return c.typ
}

func (c *stubComputed) Zone() string {
	return c.zone
}

func (c *stubComputed) Instruction() handler.ResourceInstructions {
	return c.instruction
}

func (c *stubComputed) SetupGracePeriod() int {
	return c.setupGracePeriod
}

func (c *stubComputed) Current() *handler.Resource {
	return c.current
}

func (c *stubComputed) Desired() *handler.Resource {
	return c.desired
}
