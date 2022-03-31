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
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/test"
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
			r: ResourceDefinitions{
				&ResourceDefServer{
					ResourceDefBase: &ResourceDefBase{
						TypeName:            "Server",
						SetupGracePeriodSec: 30,
					},
					Selector: &MultiZoneSelector{
						ResourceSelector: &ResourceSelector{
							Names: []string{"test-name"},
						},
						Zones: []string{"is1a"},
					},
					DedicatedCPU: true,
					ParentDef: &ParentResourceDef{
						TypeName: "EnhancedLoadBalancer",
						Selector: &NameOrSelector{ResourceSelector{Names: []string{"test-name"}}},
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
			},
			args: args{
				data: []byte(`
- type: Server
  selector:
    names: ["test-name"]
    zones: ["is1a"]
  dedicated_cpu: true
  parent:
    type: ELB
    selector: "test-name"
  setup_grace_period: 30
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
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "stub1"}},
			},
			want: nil,
		},
		{
			name: "duplicated",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "duplicated"}},
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "duplicated"}},
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
