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

type ServerGroupPlan struct {
	Name string `yaml:"name"`
	Size int    `yaml:"size"`
}

func (p *ServerGroupPlan) PlanName() string {
	return p.Name
}
func (p *ServerGroupPlan) Equals(resource interface{}) bool {
	size, ok := resource.(int)
	if !ok {
		return false
	}
	return size == p.Size
}
func (p *ServerGroupPlan) LessThan(resource interface{}) bool {
	size, ok := resource.(int)
	if !ok {
		return false
	}
	return p.Size < size
}

func (p *ServerGroupPlan) LessThanPlan(plan ResourcePlan) bool {
	sgPlan, ok := plan.(*ServerGroupPlan)
	if !ok {
		return false
	}
	return p.Size < sgPlan.Size
}
