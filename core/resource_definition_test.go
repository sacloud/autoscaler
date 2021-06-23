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

	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func TestResourceSelector_Validate(t *testing.T) {
	type fields struct {
		ID    types.ID
		Names []string
		Zones []string
	}
	type args struct {
		requireZone bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "returns error when selector is empty",
			fields:  fields{},
			args:    args{requireZone: false},
			wantErr: true,
		},
		{
			name: "returns error when both of ID and Names are specified",
			fields: fields{
				ID:    1,
				Names: []string{"1"},
			},
			args:    args{requireZone: false},
			wantErr: true,
		},
		{
			name: "returns error when zones are specified with requireZone:false",
			fields: fields{
				ID:    1,
				Zones: []string{"is1a"},
			},
			args:    args{requireZone: false},
			wantErr: true,
		},
		{
			name: "returns no error when zones are specified with requireZone:true",
			fields: fields{
				ID:    1,
				Zones: []string{"is1a"},
			},
			args:    args{requireZone: true},
			wantErr: false,
		},
		{
			name: "returns error when invalid zone value is specified",
			fields: fields{
				ID:    1,
				Zones: []string{"invalid"},
			},
			args:    args{requireZone: true},
			wantErr: true,
		},
		{
			name: "returns error when empty zone name is specified",
			fields: fields{
				ID:    1,
				Zones: []string{"invalid"},
			},
			args:    args{requireZone: true},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ResourceSelector{
				ID:    tt.fields.ID,
				Names: tt.fields.Names,
				Zones: tt.fields.Zones,
			}
			if err := rs.Validate(tt.args.requireZone); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
