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

	"github.com/sacloud/autoscaler/defaults"
	"github.com/stretchr/testify/require"
)

func TestCore_ResourceName(t *testing.T) {
	tests := []struct {
		name      string
		resources ResourceDefinitions
		args      string
		want      string
		wantErr   bool
	}{
		{
			name: "empty resource name with a definition",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name1"},
				},
			},
			args:    "",
			want:    "name1",
			wantErr: false,
		},
		{
			name: "default resource name with a definition",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name1"},
				},
			},
			args:    defaults.ResourceName,
			want:    "name1",
			wantErr: false,
		},
		{
			name: "empty resource name with definitions",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name1"},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name2"},
				},
			},
			args:    "",
			want:    "",
			wantErr: true,
		},
		{
			name: "default resource name with definitions",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name1"},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name2"},
				},
			},
			args:    defaults.ResourceName,
			want:    "",
			wantErr: true,
		},
		{
			name: "not exist name with definitions",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name1"},
				},
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{TypeName: "stub", DefName: "name2"},
				},
			},
			args:    "name3",
			want:    "name3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Core{
				listenAddress: defaults.CoreSocketAddr,
				config: &Config{
					Resources: tt.resources,
				},
			}
			got, err := c.ResourceName(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
