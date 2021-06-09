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

import "github.com/sacloud/libsacloud/v2/sacloud"

type RouterPlan struct {
	Name      string `yaml:"name"`
	BandWidth int    `yaml:"band_width"`
}

func (p *RouterPlan) PlanName() string {
	return p.Name
}
func (p *RouterPlan) Equals(resource interface{}) bool {
	v, ok := resource.(*sacloud.Internet)
	if !ok {
		return false
	}
	return v.BandWidthMbps == p.BandWidth
}
func (p *RouterPlan) LessThan(resource interface{}) bool {
	v, ok := resource.(*sacloud.Internet)
	if !ok {
		return false
	}
	return p.BandWidth < v.BandWidthMbps
}

func (p *RouterPlan) GreaterThan(resource interface{}) bool {
	v, ok := resource.(*sacloud.Internet)
	if !ok {
		return false
	}
	return v.BandWidthMbps < p.BandWidth
}

func (p *RouterPlan) LessThanPlan(plan ResourcePlan) bool {
	elbPlan, ok := plan.(*RouterPlan)
	if !ok {
		return false
	}
	return p.BandWidth < elbPlan.BandWidth
}
