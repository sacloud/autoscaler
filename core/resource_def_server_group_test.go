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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/stretchr/testify/require"
)

func TestResourceDefServerGroup_Compute(t *testing.T) {
	server1, cleanup1 := test.AddTestServer(t, "resource-def-server-test-1")
	server2, cleanup2 := test.AddTestServer(t, "resource-def-server-test-2")
	defer cleanup1()
	defer cleanup2()

	server1.CreatedAt = time.Time{}
	server2.CreatedAt = time.Time{}

	type args struct {
		ctx *RequestContext
	}
	tests := []struct {
		name    string
		def     *ResourceDefServerGroup
		args    args
		want    Resources
		wantErr bool
	}{
		{
			name: "minimum",
			def: &ResourceDefServerGroup{
				Name:    "autoscaler",
				Zone:    test.Zone,
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:       requestTypeUp,
					source:            "default",
					action:            "default",
					resourceGroupName: "default",
				}, nil, test.Logger),
			},
			want: Resources{
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &sacloud.Server{
						Name:                 "autoscaler-001",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "scale up",
			def: &ResourceDefServerGroup{
				Name:    "resource-def-server-test",
				Zone:    test.Zone,
				MinSize: 1,
				MaxSize: 3,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:       requestTypeUp,
					source:            "default",
					action:            "default",
					resourceGroupName: "default",
				}, nil, test.Logger),
			},
			want: Resources{
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server:       server1,
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_NOOP,
					indexInGroup: 0,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server:       server2,
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_NOOP,
					indexInGroup: 1,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &sacloud.Server{
						Name:                 "resource-def-server-test-003",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "scale down",
			def: &ResourceDefServerGroup{
				Name:    "resource-def-server-test",
				Zone:    test.Zone,
				MinSize: 1,
				MaxSize: 3,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:       requestTypeDown,
					source:            "default",
					action:            "default",
					resourceGroupName: "default",
				}, nil, test.Logger),
			},
			want: Resources{
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server:       server1,
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_NOOP,
					indexInGroup: 0,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server:       server2,
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_DELETE,
					indexInGroup: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.def.Compute(tt.args.ctx, test.APIClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, resource := range got {
				r, ok := resource.(*ResourceServerGroupInstance)
				if !ok {
					t.Errorf("got invalid resource type: %+#v", r)
				}
				r.def = nil // 後で比較するときのため
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestResourceDefServerGroup_Validate(t *testing.T) {
	tests := []struct {
		name string
		def  *ResourceDefServerGroup
		want []error
	}{
		{
			name: "empty",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
			},
			want: []error{
				fmt.Errorf("name: required"),
				fmt.Errorf("zone: required"),
				fmt.Errorf("min_size: required"),
				fmt.Errorf("max_size: required"),
				fmt.Errorf("template: required"),
			},
		},
		{
			name: "min/mas size",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
				Name:    "test",
				Zone:    "is1a",
				MinSize: 2,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: []error{
				fmt.Errorf("min_size: ltefield=MaxSize"),
				fmt.Errorf("max_size: gtecsfield=MinSize"),
			},
		},
		{
			name: "minimum valid definition",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
				Name:    "test",
				Zone:    "is1a",
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.def.Validate(testContext(), test.APIClient)
			require.EqualValues(t, tt.want, got)
		})
	}
}
