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
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
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
			name: "omit resource name",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB"}},
			},
			want: nil,
		},
		{
			name: "omit resource name with multiple resources",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB"}},
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "stub1"}},
			},
			want: []error{
				validate.Errorf("name is required if the configuration has more than one resource"),
			},
		},
		{
			name: "duplicated",
			rds: ResourceDefinitions{
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "duplicated"}},
				&stubResourceDef{ResourceDefBase: &ResourceDefBase{TypeName: "ELB", DefName: "duplicated"}},
			},
			want: []error{
				validate.Errorf("resource name duplicated is duplicated"),
			},
		},
		{
			name: "call definition's Validate() only when it passes structure validation",
			rds: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "ELB"},
					Dummy:           "dummy",
					validateFunc: func(ctx context.Context, apiClient iaas.APICaller) []error {
						return []error{fmt.Errorf("xxx")}
					},
				},
			},
			want: []error{
				multierror.Prefix(validate.Struct(&stubResourceDef{Dummy: "dummy"}), "resource=EnhancedLoadBalancer"),
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

func TestResourceDefinitions_LastModifiedAt(t *testing.T) {
	tests := []struct {
		name    string
		rds     ResourceDefinitions
		want    time.Time
		wantErr bool
	}{
		{
			name:    "empty",
			rds:     ResourceDefinitions{},
			want:    time.Time{},
			wantErr: false,
		},
		{
			name: "returns last modified_at",
			rds: ResourceDefinitions{
				&stubResourceDef{lastModifiedAt: time.UnixMilli(100)},
				&stubResourceDef{lastModifiedAt: time.UnixMilli(300)},
				&stubResourceDef{lastModifiedAt: time.UnixMilli(200)},
			},
			want:    time.UnixMilli(300),
			wantErr: false,
		},
		{
			name: "returns error",
			rds: ResourceDefinitions{
				&stubResourceDef{lastModifiedAt: time.UnixMilli(100)},
				&stubResourceDef{lastmodifiedAtErr: fmt.Errorf("dummy")},
				&stubResourceDef{lastModifiedAt: time.UnixMilli(200)},
			},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rds.LastModifiedAt(nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("LastModifiedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastModifiedAt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
