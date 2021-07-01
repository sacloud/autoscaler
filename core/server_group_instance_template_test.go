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

func TestServerGroupInstanceTemplate_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  *ServerGroupInstanceTemplate
	}{
		{
			name: "invalid",
			args: args{
				data: []byte("invalid"),
			},
			wantErr: true,
		},
		{
			name: "minimum",
			args: args{
				data: []byte(`
plan:
  core: 1
  memory: 1
`),
			},
			expect: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s ServerGroupInstanceTemplate
			if err := yaml.UnmarshalWithOptions(tt.args.data, &s, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expect != nil {
				require.EqualValues(t, tt.expect, &s)
			}
		})
	}
}

func TestServerGroupNICUpstream_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  *ServerGroupNICUpstream
	}{
		{
			name: "shared",
			args: args{
				data: []byte("shared"),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:      []byte("shared"),
				shared:   true,
				selector: nil,
			},
		},
		{
			name: "selector",
			args: args{
				data: []byte(`names: ["test"]`),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:    []byte(`names: ["test"]`),
				shared: false,
				selector: &ResourceSelector{
					Names: []string{"test"},
				},
			},
		},
		{
			name: "invalid",
			args: args{
				data: []byte("foobar"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s ServerGroupNICUpstream
			if err := yaml.UnmarshalWithOptions(tt.args.data, &s, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expect != nil {
				require.EqualValues(t, tt.expect, &s)
			}
		})
	}
}

func TestServerGroupNICTemplate_IPAddressByIndexFromCidrBlock(t1 *testing.T) {
	type fields struct {
		Upstream        *ServerGroupNICUpstream
		AssignCidrBlock string
	}
	type args struct {
		index int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name: "shared",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{shared: true},
			},
			args: args{
				index: 0,
			},
			want:    "",
			want1:   -1,
			wantErr: false,
		},
		{
			name: "basic switched network",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 0,
			},
			want:    "192.168.0.1",
			want1:   24,
			wantErr: false,
		},
		{
			name: "basic switched network second address",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 1,
			},
			want:    "192.168.0.2",
			want1:   24,
			wantErr: false,
		},
		{
			name: "basic switched network with carry error",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 256,
			},
			want:    "",
			want1:   -1,
			wantErr: true,
		},
		{
			name: "basic switched network without carry error",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/16",
			},
			args: args{
				index: 256,
			},
			want:    "192.168.1.1",
			want1:   16,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ServerGroupNICTemplate{
				Upstream:        tt.fields.Upstream,
				AssignCidrBlock: tt.fields.AssignCidrBlock,
			}
			got, got1, err := t.IPAddressByIndexFromCidrBlock(tt.args.index)
			if (err != nil) != tt.wantErr {
				t1.Errorf("IPAddressByIndexFromCidrBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("IPAddressByIndexFromCidrBlock() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("IPAddressByIndexFromCidrBlock() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}