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
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/stub"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestResourceDefinitions_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		r       ResourceDefinitions
		args    args
		wantErr bool
	}{
		{
			name: "unknown key",
			r:    nil,
			args: args{
				data: []byte(`
- type: Server
  selector:
    names: ["test-name"]
    zones: ["is1a"]
  unknown_key: "foobar"
`),
			},
			wantErr: true,
		},
		{
			name: "resource group",
			r: func() ResourceDefinitions {
				dns := &ResourceDefDNS{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "DNS",
					},
					Selector: &ResourceSelector{
						Names: []string{"test-name"},
					},
				}
				childServer := &ResourceDefServer{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "Server",
					},
					Selector: &MultiZoneSelector{
						ResourceSelector: &ResourceSelector{
							Names: []string{"test-child"},
						},
						Zones: []string{"is1a"},
					},
				}
				childServer.SetParent(dns)
				dns.children = ResourceDefinitions{childServer}

				return ResourceDefinitions{
					&ResourceDefServer{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "Server",
						},
						Selector: &MultiZoneSelector{
							ResourceSelector: &ResourceSelector{
								Names: []string{"test-name"},
							},
							Zones: []string{"is1a"},
						},
						DedicatedCPU: true,
					},
					dns,
					&ResourceDefGSLB{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "GSLB",
						},
						Selector: &ResourceSelector{
							Names: []string{"test-name"},
						},
					},
					&ResourceDefELB{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "EnhancedLoadBalancer",
						},
						Selector: &ResourceSelector{
							Names: []string{"test-name"},
						},
					},
				}
			}(),
			args: args{
				data: []byte(`
- type: Server
  selector:
    names: ["test-name"]
    zones: ["is1a"]
  dedicated_cpu: true
- type: DNS
  selector:
    names: ["test-name"]
  resources:
    - type: Server
      selector:
        names: ["test-child"]
        zones: ["is1a"]
- type: GSLB 
  selector:
    names: ["test-name"]
- type: ELB
  selector:
    names: ["test-name"]
`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target ResourceDefinitions
			if err := yaml.UnmarshalWithOptions(tt.args.data, &target, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, tt.r, target)
		})
	}
}

func TestResourceDefinitions_HandleAll_havingChildrenDefinitionReturnsMultipleResource(t *testing.T) {
	ctx := testContext()
	defs := ResourceDefinitions{
		&stubResourceDef{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "stub",
				children: ResourceDefinitions{
					&stubResourceDef{},
				},
			},
			computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
				return Resources{
					&stubResource{
						ResourceBase: &ResourceBase{resourceType: ResourceTypeUnknown},
						computeFunc: func(ctx *RequestContext, refresh bool) (Computed, error) {
							return &stubComputed{}, nil
						},
					},
					&stubResource{
						ResourceBase: &ResourceBase{resourceType: ResourceTypeUnknown},
						computeFunc: func(ctx *RequestContext, refresh bool) (Computed, error) {
							return &stubComputed{}, nil
						},
					},
				}, nil
			},
		},
	}
	err := defs.handleAll(ctx, test.APIClient, noopHandlers, nil, defs)
	require.True(t, err != nil)
}

func TestResourceDefinitions_HandleAll_withActualResource(t *testing.T) {
	ctx := testContext()
	_, cleanup := test.AddTestServer(t, "test-server")
	defer cleanup()
	_, cleanup2 := test.AddTestDNS(t, "test-dns.com")
	defer cleanup2()

	server := &ResourceDefServer{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "Server",
		},
		Selector: &MultiZoneSelector{
			ResourceSelector: &ResourceSelector{
				Names: []string{"test-server"},
			},
			Zones: []string{test.Zone},
		},
	}
	dns := &ResourceDefDNS{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "DNS",
			children: ResourceDefinitions{server},
		},
		Selector: &ResourceSelector{
			Names: []string{"test-dns.com"},
		},
	}
	defs := ResourceDefinitions{dns}

	var called []string
	stubHandler := &Handler{
		Name: "stub",
		BuiltinHandler: &stub.Handler{
			Logger: test.Logger,
			HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
				if server := request.Desired.GetServer(); server != nil {
					// HandleAllの中でParentが設定されているか
					require.NotNil(t, server.Parent.GetDns())

					called = append(called, "server")
				} else if dns := request.Desired.GetDns(); dns != nil {
					called = append(called, "dns")
				}
				return nil
			},
		},
	}

	err := defs.handleAll(ctx, test.APIClient, Handlers{stubHandler}, nil, defs)
	require.NoError(t, err)
	// 子から先にHandleされているか?
	require.Equal(t, []string{"server", "dns"}, called)
}

var noopHandlers = Handlers{
	&Handler{
		Name: "stub",
		BuiltinHandler: &stub.Handler{
			Logger: test.Logger,
			HandleFunc: func(_ *handler.HandleRequest, _ handlers.ResponseSender) error {
				return nil
			},
		},
	},
}

func TestResourceDefinitions_FilterByResourceName(t *testing.T) {
	tests := []struct {
		name         string
		rds          ResourceDefinitions
		resourceName string
		want         ResourceDefinitions
	}{
		{
			name: "minimum",
			rds: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test",
						children: nil,
					},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test2",
					},
				},
			},
			resourceName: "test",
			want: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test",
						children: nil,
					},
				},
			},
		},
		{
			name: "not exist",
			rds: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test",
						children: nil,
					},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test2",
					},
				},
			},
			resourceName: "not exist",
			want:         nil,
		},
		{
			name: "returns parent if child is hit",
			rds: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test",
						children: ResourceDefinitions{
							&stubResourceDef{
								ResourceDefBase: &ResourceDefBase{
									TypeName: "stub",
									DefName:  "child",
								},
							},
						},
					},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test2",
					},
				},
			},
			resourceName: "test",
			want: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "test",
						children: ResourceDefinitions{
							&stubResourceDef{
								ResourceDefBase: &ResourceDefBase{
									TypeName: "stub",
									DefName:  "child",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rds.FilterByResourceName(tt.resourceName)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestResourceDefinitions_Validate(t *testing.T) {
	tests := []struct {
		name string
		rds  ResourceDefinitions
		want []error
	}{
		{
			name: "no error",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "DNS", DefName: "stub1"}},
			},
			want: nil,
		},
		{
			name: "duplicated",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "DNS", DefName: "duplicated"}},
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "DNS", DefName: "duplicated"}},
			},
			want: []error{
				fmt.Errorf("resource name duplicated is duplicated"),
			},
		},
		{
			name: "duplicated with nested defs",
			rds: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "DNS",
						DefName:  "duplicated",
						children: ResourceDefinitions{
							&stubResourceDef{
								ResourceDefBase: &ResourceDefBase{
									TypeName: "DNS",
									DefName:  "stub1-1",
									children: ResourceDefinitions{
										&stubResourceDef{
											ResourceDefBase: &ResourceDefBase{
												TypeName: "DNS",
												DefName:  "stub1-1-1",
											},
										},
									},
								},
							},
						},
					},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "DNS",
						DefName:  "stub2",
						children: ResourceDefinitions{
							&stubResourceDef{
								ResourceDefBase: &ResourceDefBase{
									TypeName: "DNS",
									DefName:  "stub2-1",
									children: ResourceDefinitions{
										&stubResourceDef{
											ResourceDefBase: &ResourceDefBase{
												TypeName: "DNS",
												DefName:  "duplicated",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []error{
				fmt.Errorf("resource name duplicated is duplicated"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rds.Validate(testContext(), test.APIClient)
			require.EqualValues(t, tt.want, got)
		})
	}
}
