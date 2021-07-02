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

func TestResourceDefGroups_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		r       *ResourceDefGroups
		args    args
		wantErr bool
	}{
		{
			name: "unknown key",
			r:    &ResourceDefGroups{},
			args: args{
				data: []byte(`
web: 
  resources:
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
			r: func() *ResourceDefGroups {
				rgs := newResourceDefGroups()
				rg := &ResourceDefGroup{
					Actions: Actions{
						"foobar": []string{"handler1", "handler2"},
					},
				}

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

				rg.ResourceDefs = ResourceDefinitions{
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
			var target ResourceDefGroups
			if err := yaml.UnmarshalWithOptions(tt.args.data, &target, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, tt.r, &target)
		})
	}
}
