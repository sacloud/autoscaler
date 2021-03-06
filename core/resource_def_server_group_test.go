// Copyright 2021-2022 The sacloud/autoscaler Authors
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
					requestType:  requestTypeUp,
					source:       "default",
					resourceName: "autoscaler",
				}, nil, test.Logger),
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
					requestType:  requestTypeUp,
					source:       "default",
					resourceName: "resource-def-server-test",
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
					requestType:  requestTypeDown,
					source:       "default",
					resourceName: "resource-def-server-test",
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
				Zone:    test.Zone,
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
				Zone:    test.Zone,
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
				Zone:    test.Zone,
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
				}, nil, test.Logger),
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
				Zone:    test.Zone,
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
				}, nil, test.Logger),
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
				Zone:    test.Zone,
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
				r.def = nil // ?????????????????????????????????
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
				Zone:    "is1a",
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
			name: "returns no error without server_name_prefix",
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
				Zone:             "is1a",
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
			name: "up without plans / with invalid server count", // max_size????????????????????????????????????????????????????????????
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
			name: "down without plans / with invalid server state", // max_size???????????????????????????????????????down???max_size?????????
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
			d.printWarningForServerNamePrefix(ctx) // nolint
			require.Equal(t, tt.wantWarn, len(writer.Bytes()) > 0)
		})
	}
}
