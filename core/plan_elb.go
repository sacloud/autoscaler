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

import "github.com/sacloud/iaas-api-go"

type ELBPlan struct {
	Name string `yaml:"name"`
	CPS  int    `yaml:"cps"`
}

func (p *ELBPlan) PlanName() string {
	return p.Name
}
func (p *ELBPlan) Equals(resource interface{}) bool {
	v, ok := resource.(*iaas.ProxyLB)
	if !ok {
		return false
	}
	return v.Plan.Int() == p.CPS
}
func (p *ELBPlan) LessThan(resource interface{}) bool {
	v, ok := resource.(*iaas.ProxyLB)
	if !ok {
		return false
	}
	return p.CPS < v.Plan.Int()
}

func (p *ELBPlan) LessThanPlan(plan ResourcePlan) bool {
	elbPlan, ok := plan.(*ELBPlan)
	if !ok {
		return false
	}
	return p.CPS < elbPlan.CPS
}
