// Copyright 2021-2023 The sacloud/autoscaler Authors
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
	"github.com/sacloud/iaas-api-go/types"
)

func computedToParents(parentComputed Computed) *handler.Parent {
	if parentComputed != nil {
		current := parentComputed.Current()
		if current != nil {
			var parent *handler.Parent

			// NOTE: Parentになれるリソースが増えた場合はここを修正する
			if v := current.GetDns(); v != nil {
				parent = &handler.Parent{
					Resource: &handler.Parent_Dns{
						Dns: v,
					},
				}
			}
			if v := current.GetElb(); v != nil {
				parent = &handler.Parent{
					Resource: &handler.Parent_Elb{
						Elb: v,
					},
				}
			}
			if v := current.GetGslb(); v != nil {
				parent = &handler.Parent{
					Resource: &handler.Parent_Gslb{
						Gslb: v,
					},
				}
			}
			if v := current.GetLoadBalancer(); v != nil {
				parent = &handler.Parent{
					Resource: &handler.Parent_LoadBalancer{
						LoadBalancer: v,
					},
				}
			}

			return parent
		}
	}
	return nil
}

func boolToCommitment(dedicatedCPU bool) types.ECommitment {
	v := types.Commitments.Standard
	if dedicatedCPU {
		v = types.Commitments.DedicatedCPU
	}
	return v
}
