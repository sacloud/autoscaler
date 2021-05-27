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

// Code generated by "stringer -type=ResourceTypes"; DO NOT EDIT.

package core

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ResourceTypeUnknown-0]
	_ = x[ResourceTypeServer-1]
	_ = x[ResourceTypeServerGroup-2]
	_ = x[ResourceTypeEnhancedLoadBalancer-3]
	_ = x[ResourceTypeGSLB-4]
	_ = x[ResourceTypeDNS-5]
}

const _ResourceTypes_name = "ResourceTypeUnknownResourceTypeServerResourceTypeServerGroupResourceTypeEnhancedLoadBalancerResourceTypeGSLBResourceTypeDNS"

var _ResourceTypes_index = [...]uint8{0, 19, 37, 60, 92, 108, 123}

func (i ResourceTypes) String() string {
	if i < 0 || i >= ResourceTypes(len(_ResourceTypes_index)-1) {
		return "ResourceTypes(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ResourceTypes_name[_ResourceTypes_index[i]:_ResourceTypes_index[i+1]]
}