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
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
	"github.com/stretchr/testify/require"
)

func TestResourceDefServerGroup_Compute(t *testing.T) {
	server1, cleanup1 := test.AddTestServer(t, "resource-def-server-test-001")
	server2, cleanup2 := test.AddTestServer(t, "resource-def-server-test-003")
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
				ResourceDefBase: &ResourceDefBase{
					DefName:  "autoscaler",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
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
					requestType:  requestTypeUp,
					source:       "default",
					resourceName: "autoscaler",
				}, test.Logger),
			},
			want: Resources{
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
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
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
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
					requestType:  requestTypeUp,
					source:       "default",
					resourceName: "resource-def-server-test",
				}, test.Logger),
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
					indexInGroup: 2,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
						Name:                 "resource-def-server-test-002",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "scale down",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
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
					requestType:  requestTypeDown,
					source:       "default",
					resourceName: "resource-def-server-test",
				}, test.Logger),
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
					indexInGroup: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "scale up with named plans",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
				MinSize: 1,
				MaxSize: 5,
				Plans: []*ServerGroupPlan{
					{Size: 1, Name: "smallest"},
					{Size: 5, Name: "largest"},
				},
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeUp,
					source:           "default",
					resourceName:     "resource-def-server-test",
					desiredStateName: "largest",
				}, test.Logger),
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
					indexInGroup: 2,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
						Name:                 "resource-def-server-test-002",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 1,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
						Name:                 "resource-def-server-test-004",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 3,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
						Name:                 "resource-def-server-test-005",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
					indexInGroup: 4,
				},
			},
			wantErr: false,
		},
		{
			name: "scale down with named plans",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
				MinSize: 1,
				MaxSize: 5,
				Plans: []*ServerGroupPlan{
					{Size: 1, Name: "smallest"},
					{Size: 5, Name: "largest"},
				},
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeDown,
					source:           "default",
					resourceName:     "resource-def-server-test",
					desiredStateName: "smallest",
				}, test.Logger),
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
					indexInGroup: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "scale up without valid named plan",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
				MinSize: 1,
				MaxSize: 5,
				Plans: []*ServerGroupPlan{
					{Size: 1, Name: "smallest"},
					{Size: 5, Name: "largest"},
				},
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeUp,
					source:           "default",
					resourceName:     "resource-def-server-test",
					desiredStateName: "smallest",
				}, test.Logger),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "scale down without valid named plan",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
				MinSize: 1,
				MaxSize: 5,
				Plans: []*ServerGroupPlan{
					{Size: 3, Name: "medium"},
					{Size: 5, Name: "largest"},
				},
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeDown,
					source:           "default",
					resourceName:     "resource-def-server-test",
					desiredStateName: "medium",
				}, test.Logger),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "scale up with named plans without desired state name",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "resource-def-server-test",
					TypeName: "ServerGroup",
				},
				Zones:   []string{test.Zone},
				MinSize: 1,
				MaxSize: 5,
				Plans: []*ServerGroupPlan{
					{Size: 1, Name: "smallest"},
					{Size: 5, Name: "largest"},
				},
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeUp,
					source:           "default",
					resourceName:     "resource-def-server-test",
					desiredStateName: "default",
				}, test.Logger),
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
					indexInGroup: 2,
				},
				&ResourceServerGroupInstance{
					ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
					apiClient:    test.APIClient,
					server: &iaas.Server{
						Name:                 "resource-def-server-test-002",
						CPU:                  1,
						MemoryMB:             1 * size.GiB,
						ServerPlanCommitment: types.Commitments.Standard,
					},
					zone:         test.Zone,
					instruction:  handler.ResourceInstructions_CREATE,
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
	validate.InitValidatorAlias(iaas.SakuraCloudZones)
	tests := []struct {
		name string
		def  *ResourceDefServerGroup
		want []error
	}{
		{
			name: "returns error when name and server_name_prefix is empty",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
				Zones:   []string{"is1a"},
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: []error{
				fmt.Errorf("resource=ServerGroup name or server_name_prefix: required"),
			},
		},
		{
			name: "returns error with invalid min/max size",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zones:   []string{"is1a"},
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
			name: "returns no error without server_name_prefix",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zones:   []string{"is1a"},
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
		{
			name: "returns no error without DefName",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
				ServerNamePrefix: "test",
				Zones:            []string{"is1a"},
				MinSize:          1,
				MaxSize:          1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: nil,
		},
		{
			name: "return error with zone and zones",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
				},
				ServerNamePrefix: "test",
				MinSize:          1,
				MaxSize:          1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
				Zone:  "is1a",
				Zones: []string{"is1a", "is1b"},
			},
			want: []error{
				fmt.Errorf("resource=ServerGroup only one of zone and zones can be specified"),
			},
		},
		{
			name: "returns only not found error when parent is LB and has zone",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zone:    "is1a",
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
				ParentDef: &ParentResourceDef{
					TypeName: ResourceTypeLoadBalancer.String(),
					Selector: &NameOrSelector{
						ResourceSelector: ResourceSelector{
							Names: []string{"foobar"},
						},
					},
				},
			},
			want: []error{
				fmt.Errorf("resource=ServerGroup resource=LoadBalancer resource not found with selector: ID: , Names: [foobar], Tags: []"),
			},
		},
		{
			name: "returns error when parent is LB and has multiple zones",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zones:   []string{"is1a", "is1b"},
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
				ParentDef: &ParentResourceDef{
					TypeName: ResourceTypeLoadBalancer.String(),
					Selector: &NameOrSelector{
						ResourceSelector: ResourceSelector{
							Names: []string{"foobar"},
						},
					},
				},
			},
			want: []error{
				fmt.Errorf("resource=ServerGroup multiple zones cannot be specified when the parent is a LoadBalancer"),
				fmt.Errorf("resource=ServerGroup resource=LoadBalancer resource not found with selector: ID: , Names: [foobar], Tags: []"),
			},
		},
		{
			name: "returns error with empty zone name",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zones:   []string{"is1a", "", "is1b"}, // 空文字
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: []error{
				fmt.Errorf("zones[1]: required"),
			},
		},
		{
			name: "returns error with duplicated zone name",
			def: &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  "test",
				},
				Zones:   []string{"is1a", "is1a", "is1b"}, // 重複
				MinSize: 1,
				MaxSize: 1,
				Template: &ServerGroupInstanceTemplate{
					Plan: &ServerGroupInstancePlan{
						Core:   1,
						Memory: 1,
					},
				},
			},
			want: []error{
				fmt.Errorf("zones: unique"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validate.StructWithMultiError(tt.def)
			if len(got) == 0 {
				got = tt.def.Validate(testContext(), test.APIClient)
			}
			require.Equal(t, len(tt.want), len(got))
			for i := range got {
				require.Equal(t, tt.want[i].Error(), got[i].Error())
			}
		})
	}
}

func TestResourceDefServerGroup_determineServerName(t *testing.T) {
	tests := []struct {
		name      string
		defName   string
		resources Resources
		want      string
		wantIndex int
	}{
		{
			name:      "from empty",
			defName:   "prefix",
			resources: nil,
			want:      "prefix-001",
			wantIndex: 0,
		},
		{
			name:    "from servers are exist",
			defName: "prefix",
			resources: Resources{
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-001"},
				},
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-002"},
				},
			},
			want:      "prefix-003",
			wantIndex: 2,
		},
		{
			name:    "servers that are not sequentially numbered",
			defName: "prefix",
			resources: Resources{
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-001"},
				},
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-003"},
				},
			},
			want:      "prefix-002",
			wantIndex: 1,
		},
		{
			name:    "exist multiple unnumbered",
			defName: "prefix",
			resources: Resources{
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-001"},
				},
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-003"},
				},
				&ResourceServerGroupInstance{
					server: &iaas.Server{Name: "prefix-005"},
				},
			},
			want:      "prefix-002",
			wantIndex: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName: tt.defName,
				},
			}
			got, index := d.determineServerName(tt.resources)
			require.EqualValues(t, tt.want, got)
			require.EqualValues(t, tt.wantIndex, index)
		})
	}
}

func TestResourceDefServerGroup_desiredPlan(t *testing.T) {
	type fields struct {
		MinSize int
		MaxSize int
		Plans   []*ServerGroupPlan
	}
	type args struct {
		ctx          *RequestContext
		currentCount int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ServerGroupPlan
		wantErr bool
	}{
		{
			name: "up: currentCount has smaller value than min_size",
			fields: fields{
				MinSize: 1,
				MaxSize: 3,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "up: currentCount has larger value than max_size",
			fields: fields{
				MinSize: 1,
				MaxSize: 3,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 4,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 4,
			},
			wantErr: false,
		},
		{
			name: "down with resources that are outside the scope of the plans:max",
			fields: fields{
				MinSize: 1,
				MaxSize: 3,
				Plans:   nil,
			},
			args: args{
				ctx:          testContextDown(),
				currentCount: 4,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 3,
			},
			wantErr: false,
		},
		{
			name: "down with resources that are outside the scope of the plans:min",
			fields: fields{
				MinSize: 1,
				MaxSize: 3,
				Plans:   nil,
			},
			args: args{
				ctx:          testContextDown(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "up without plans / without servers on cloud",
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "up without plans / with servers on cloud",
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 1,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "up without plans / with invalid server count", // max_sizeを超えたサーバがある場合は特に何もしない
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 2,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 2,
			},
			wantErr: false,
		},
		{
			name: "down without plans / without servers on cloud",
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContextDown(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "down without plans / with servers on cloud",
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContextDown(),
				currentCount: 1,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "down without plans / with invalid server state", // max_sizeを超えたサーバがある場合のdownはmax_sizeを返す
			fields: fields{
				MinSize: 0,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContextDown(),
				currentCount: 2,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "up with same min/max size",
			fields: fields{
				MinSize: 1,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
		{
			name: "down with same min/max size",
			fields: fields{
				MinSize: 1,
				MaxSize: 1,
				Plans:   nil,
			},
			args: args{
				ctx:          testContext(),
				currentCount: 0,
			},
			want: &ServerGroupPlan{
				Name: "",
				Size: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName:  "default",
					TypeName: "ServerGroup",
				},
				MinSize: tt.fields.MinSize,
				MaxSize: tt.fields.MaxSize,
				Plans:   tt.fields.Plans,
			}

			got, err := d.desiredPlan(tt.args.ctx, tt.args.currentCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredPlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestResourceDefServerGroup_filterCloudServers(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		servers []*iaas.Server
		want    []*iaas.Server
	}{
		{
			name:   "minimum",
			prefix: "foo",
			servers: []*iaas.Server{
				{Name: "foo-001"},
				{Name: "foo-002"},
				{Name: "bar-001"},
			},
			want: []*iaas.Server{
				{Name: "foo-001"},
				{Name: "foo-002"},
			},
		},
		{
			name:   "filtered by prefix",
			prefix: "bar",
			servers: []*iaas.Server{
				{Name: "bar-001"},
				{Name: "bar-002"},
				{Name: "foobar-001"},
			},
			want: []*iaas.Server{
				{Name: "bar-001"},
				{Name: "bar-002"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					TypeName: "ServerGroup",
					DefName:  tt.prefix,
				},
				ServerNamePrefix: tt.prefix,
			}
			require.Equal(t, tt.want, d.filterCloudServers(tt.servers))
		})
	}
}

func TestResourceDefServerGroup_printWarningForServerNamePrefix(t *testing.T) {
	writer := new(bytes.Buffer)
	ctx := config.NewLoadConfigContext(
		context.Background(), false, log.NewLogger(&log.LoggerOption{Writer: writer}),
	)
	type fields struct {
		Name             string
		ServerNamePrefix string
	}
	tests := []struct {
		name     string
		fields   fields
		wantWarn bool
	}{
		{
			name: "warn with only name",
			fields: fields{
				Name:             "foo",
				ServerNamePrefix: "",
			},
			wantWarn: true,
		},
		{
			name: "no warn with server_name_prefix",
			fields: fields{
				Name:             "",
				ServerNamePrefix: "foo",
			},
			wantWarn: false,
		},
		{
			name: "no warn with both of name and server_name_prefix",
			fields: fields{
				Name:             "foo",
				ServerNamePrefix: "foo",
			},
			wantWarn: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.Reset()

			d := &ResourceDefServerGroup{
				ResourceDefBase:  &ResourceDefBase{DefName: tt.fields.Name},
				ServerNamePrefix: tt.fields.ServerNamePrefix,
			}
			d.printWarningForServerNamePrefix(ctx) //nolint
			require.Equal(t, tt.wantWarn, len(writer.Bytes()) > 0)
		})
	}
}

func TestResourceDefServerGroup_determineZone(t *testing.T) {
	tests := []struct {
		name  string
		zones []string
		index int
		want  string
	}{
		{
			name:  "minimum",
			zones: []string{"is1a"},
			index: 0,
			want:  "is1a",
		},
		{
			name:  "index is greater than zones",
			zones: []string{"is1a"},
			index: 1,
			want:  "is1a",
		},
		{
			name:  "multiple zones",
			zones: []string{"is1a", "tk1a"},
			index: 0,
			want:  "is1a",
		},
		{
			name:  "multiple zones",
			zones: []string{"is1a", "tk1a"},
			index: 1,
			want:  "tk1a",
		},
		{
			name:  "index is greater than zones with multiple zones",
			zones: []string{"is1a", "tk1a"},
			index: 2,
			want:  "is1a",
		},
		{
			name:  "index is greater than zones with multiple zones",
			zones: []string{"is1a", "tk1a"},
			index: 3,
			want:  "tk1a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefServerGroup{Zones: tt.zones}
			if got := d.determineZone(tt.index); got != tt.want {
				t.Errorf("determineZone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceDefServerGroup_lastModifiedAt(t *testing.T) {
	type args struct {
		cloudResources []*iaas.Server
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "empty",
			args: args{},
			want: time.Time{},
		},
		{
			name: "returns modified-at simply",
			args: args{
				cloudResources: []*iaas.Server{
					{ModifiedAt: time.UnixMilli(100)},
				},
			},
			want: time.UnixMilli(100),
		},
		{
			name: "returns last modified-at",
			args: args{
				cloudResources: []*iaas.Server{
					{ModifiedAt: time.UnixMilli(103)},
					{ModifiedAt: time.UnixMilli(107)},
					{ModifiedAt: time.UnixMilli(105)},
					{ModifiedAt: time.UnixMilli(101)},
				},
			},
			want: time.UnixMilli(107),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefServerGroup{}
			if got := d.lastModifiedAt(tt.args.cloudResources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastModifiedAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
