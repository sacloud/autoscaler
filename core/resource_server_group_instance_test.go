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
	"testing"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/iaas-api-go"
	"github.com/stretchr/testify/require"
)

func TestResourceServerGroupInstance_computeNetworkInterfaces(t *testing.T) {
	sw, cleanup := test.AddTestSwitch(t, "test-switch")
	defer cleanup()

	type fields struct {
		ResourceBase *ResourceBase
		server       *iaas.Server
		def          *ResourceDefServerGroup
		instruction  handler.ResourceInstructions
		indexInGroup int
	}
	type args struct {
		ctx *RequestContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*handler.ServerGroupInstance_NIC
		wantErr bool
	}{
		{
			name: "new server",
			fields: fields{
				ResourceBase: &ResourceBase{
					resourceType: ResourceTypeServerGroupInstance,
				},
				server: &iaas.Server{
					Name: "autoscaler-001",
				},
				def: &ResourceDefServerGroup{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "ServerGroup",
						DefName:  "autoscaler",
					},
					Zone:    test.Zone,
					MinSize: 1,
					MaxSize: 1,
					Template: &ServerGroupInstanceTemplate{
						NetworkInterfaces: []*ServerGroupNICTemplate{
							{
								Upstream: &ServerGroupNICUpstream{
									selector: &ResourceSelector{
										Names: []string{sw.Name},
									},
								},
								AssignCidrBlock:  "192.168.1.16/28",
								AssignNetMaskLen: 28,
								DefaultRoute:     "192.168.1.1",
							},
						},
					},
				},
				instruction:  handler.ResourceInstructions_CREATE,
				indexInGroup: 0,
			},
			args: args{
				ctx: testContext(),
			},
			want: []*handler.ServerGroupInstance_NIC{
				{
					Upstream:      sw.ID.String(),
					UserIpAddress: "192.168.1.17",
					AssignedNetwork: &handler.NetworkInfo{
						IpAddress: "192.168.1.17",
						Netmask:   28,
						Gateway:   "192.168.1.1",
						Index:     0,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceServerGroupInstance{
				ResourceBase: tt.fields.ResourceBase,
				apiClient:    test.APIClient,
				server:       tt.fields.server,
				zone:         test.Zone,
				def:          tt.fields.def,
				instruction:  tt.fields.instruction,
				indexInGroup: tt.fields.indexInGroup,
			}
			got, err := r.computeNetworkInterfaces(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("computeNetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestResourceServerGroupInstance_healthCheckRequests(t *testing.T) {
	type fields struct {
		ResourceBase *ResourceBase
		server       *iaas.Server
		def          *ResourceDefServerGroup
		instruction  handler.ResourceInstructions
		indexInGroup int
	}
	sw, cleanup := test.AddTestSwitch(t, "test-switch-hc")
	defer cleanup()

	tests := []struct {
		name    string
		fields  fields
		want    []*ChildResourceHealthCheckRequest
		wantErr bool
	}{
		{
			name: "no ExposeInfo",
			fields: fields{
				ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
				server:       &iaas.Server{Name: "autoscaler-001"},
				def: &ResourceDefServerGroup{
					ResourceDefBase: &ResourceDefBase{TypeName: "ServerGroup", DefName: "autoscaler"},
					Zone:            test.Zone,
					MinSize:         1,
					MaxSize:         1,
					Template: &ServerGroupInstanceTemplate{
						NetworkInterfaces: []*ServerGroupNICTemplate{{
							Upstream: &ServerGroupNICUpstream{selector: &ResourceSelector{Names: []string{sw.Name}}},
						}},
					},
				},
				instruction:  handler.ResourceInstructions_CREATE,
				indexInGroup: 0,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "empty Vips/Ports",
			fields: fields{
				ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
				server:       &iaas.Server{Name: "autoscaler-002"},
				def: &ResourceDefServerGroup{
					ResourceDefBase: &ResourceDefBase{TypeName: "ServerGroup", DefName: "autoscaler"},
					Zone:            test.Zone,
					MinSize:         1,
					MaxSize:         1,
					Template: &ServerGroupInstanceTemplate{
						NetworkInterfaces: []*ServerGroupNICTemplate{{
							Upstream:         &ServerGroupNICUpstream{selector: &ResourceSelector{Names: []string{sw.Name}}},
							ExposeInfo:       &ServerGroupNICMetadata{},
							AssignCidrBlock:  "192.168.1.16/28",
							AssignNetMaskLen: 28,
						}},
					},
				},
				instruction:  handler.ResourceInstructions_CREATE,
				indexInGroup: 0,
			},
			want: []*ChildResourceHealthCheckRequest{{
				VIP:       "",
				IPAddress: "192.168.1.17",
				Port:      0,
			}},
			wantErr: false,
		},
		{
			name: "multiple Vips/Ports",
			fields: fields{
				ResourceBase: &ResourceBase{resourceType: ResourceTypeServerGroupInstance},
				server:       &iaas.Server{Name: "autoscaler-003"},
				def: &ResourceDefServerGroup{
					ResourceDefBase: &ResourceDefBase{TypeName: "ServerGroup", DefName: "autoscaler"},
					Zone:            test.Zone,
					MinSize:         1,
					MaxSize:         1,
					Template: &ServerGroupInstanceTemplate{
						NetworkInterfaces: []*ServerGroupNICTemplate{{
							Upstream: &ServerGroupNICUpstream{selector: &ResourceSelector{Names: []string{sw.Name}}},
							ExposeInfo: &ServerGroupNICMetadata{
								VIPs:  []string{"10.0.0.1", "10.0.0.2"},
								Ports: []int{80, 443},
							},
						}},
					},
				},
				instruction:  handler.ResourceInstructions_CREATE,
				indexInGroup: 0,
			},
			want: []*ChildResourceHealthCheckRequest{
				{VIP: "10.0.0.1", IPAddress: "", Port: 80},
				{VIP: "10.0.0.2", IPAddress: "", Port: 80},
				{VIP: "10.0.0.1", IPAddress: "", Port: 443},
				{VIP: "10.0.0.2", IPAddress: "", Port: 443},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceServerGroupInstance{
				ResourceBase: tt.fields.ResourceBase,
				apiClient:    test.APIClient,
				server:       tt.fields.server,
				zone:         test.Zone,
				def:          tt.fields.def,
				instruction:  tt.fields.instruction,
				indexInGroup: tt.fields.indexInGroup,
			}
			ctx := testContext()
			got, err := r.healthCheckRequests(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("healthCheckRequests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}
