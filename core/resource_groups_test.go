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
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestResourceGroups_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		r       *ResourceGroups
		args    args
		wantErr bool
	}{
		{
			name: "resource group",
			r: func() *ResourceGroups {
				rgs := newResourceGroups()
				rg := &ResourceGroup{
					Actions: Actions{
						"foobar": []string{"handler1", "handler2"},
					},
				}

				dns := &ResourceDefDNS{
					ResourceBase: &ResourceBase{
						TypeName: "DNS",
						TargetSelector: &ResourceSelector{
							Names: []string{"test-name"},
							Zone:  "is1a",
						},
					},
				}
				childServer := &ResourceDefServer{
					ResourceBase: &ResourceBase{
						TypeName: "Server",
						TargetSelector: &ResourceSelector{
							Names: []string{"test-child"},
							Zone:  "is1a",
						},
					},
				}
				childServer.SetParent(dns)
				dns.children = ResourceDefinitions{childServer}

				rg.Resources = ResourceDefinitions{
					&ResourceDefServer{
						ResourceBase: &ResourceBase{
							TypeName: "Server",
							TargetSelector: &ResourceSelector{
								Names: []string{"test-name"},
								Zone:  "is1a",
							},
						},
						DedicatedCPU: true,
					},
					&ResourceDefServerGroup{
						ResourceBase: &ResourceBase{
							TypeName: "ServerGroup",
							TargetSelector: &ResourceSelector{
								Names: []string{"test-name"},
								Zone:  "is1a",
							},
						},
					},
					dns,
					&ResourceDefGSLB{
						ResourceBase: &ResourceBase{
							TypeName: "GSLB",
							TargetSelector: &ResourceSelector{
								Names: []string{"test-name"},
								Zone:  "is1a",
							},
						},
					},
					&ResourceDefELB{
						ResourceBase: &ResourceBase{
							TypeName: "EnhancedLoadBalancer",
							TargetSelector: &ResourceSelector{
								Names: []string{"test-name"},
								Zone:  "is1a",
							},
						},
					},
				}
				rgs.Set("web", rg)
				return rgs
			}(),
			args: args{
				data: []byte(`
web: 
  resources:
    - type: Server
      selector:
        names: ["test-name"]
        zone: "is1a"
      dedicated_cpu: true
    - type: ServerGroup
      selector:
        names: ["test-name"]
        zone: "is1a"
    - type: DNS
      selector:
        names: ["test-name"]
        zone: "is1a"
      resources:
        - type: Server
          selector:
            names: ["test-child"]
            zone: "is1a"
    - type: GSLB 
      selector:
        names: ["test-name"]
        zone: "is1a"
    - type: ELB
      selector:
        names: ["test-name"]
        zone: "is1a"
  actions:
    foobar:
      - handler1
      - handler2
`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target ResourceGroups
			if err := yaml.Unmarshal(tt.args.data, &target); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, tt.r, &target)
		})
	}
}
