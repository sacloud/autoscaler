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
	"testing"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/iaas-api-go/types"
)

func TestMultiZoneSelector_Validate(t *testing.T) {
	type fields struct {
		ID    types.ID
		Names []string
		Zones []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "returns error when selector is empty",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "returns error when both of ID and Names are specified",
			fields: fields{
				ID:    1,
				Names: []string{"1"},
			},
			wantErr: true,
		},
		{
			name: "returns error when invalid zone value is specified",
			fields: fields{
				ID:    1,
				Zones: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "returns error when empty zone name is specified",
			fields: fields{
				ID:    1,
				Zones: []string{"invalid"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &MultiZoneSelector{
				ResourceSelector: &ResourceSelector{
					ID:    tt.fields.ID,
					Names: tt.fields.Names,
				},
				Zones: tt.fields.Zones,
			}
			if err := rs.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceDefBase_SetupGracePeriod(t *testing.T) {
	type fields struct {
		TypeName            string
		DefName             string
		SetupGracePeriodSec int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "server with default",
			fields: fields{
				TypeName: "Server",
				DefName:  "test1",
			},
			want: defaults.SetupGracePeriods[ResourceTypeServer.String()],
		},
		{
			name: "other types",
			fields: fields{
				TypeName:            "ELB",
				DefName:             "test1",
				SetupGracePeriodSec: 30,
			},
			want: 30,
		},
		{
			name: "other types with default",
			fields: fields{
				TypeName: "ELB",
				DefName:  "test1",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceDefBase{
				TypeName:            tt.fields.TypeName,
				DefName:             tt.fields.DefName,
				SetupGracePeriodSec: tt.fields.SetupGracePeriodSec,
			}
			if got := r.SetupGracePeriod(); got != tt.want {
				t.Errorf("SetupGracePeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}
