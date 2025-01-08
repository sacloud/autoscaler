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

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/packages-go/size"
)

func TestServerPlan_Equals(t *testing.T) {
	type fields struct {
		Name   string
		Core   int
		Memory int
	}
	type args struct {
		resource interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "returns true",
			fields: fields{
				Name:   "",
				Core:   1,
				Memory: 1,
			},
			args: args{
				&iaas.Server{
					CPU:      1,
					MemoryMB: 1 * size.GiB,
				},
			},
			want: true,
		},
		{
			name: "returns false",
			fields: fields{
				Name:   "",
				Core:   1,
				Memory: 2,
			},
			args: args{
				&iaas.Server{
					CPU:      1,
					MemoryMB: 1 * size.GiB,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ServerPlan{
				Name:   tt.fields.Name,
				Core:   tt.fields.Core,
				Memory: tt.fields.Memory,
			}
			if got := p.Equals(tt.args.resource); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerPlan_LessThan(t *testing.T) {
	type fields struct {
		Name   string
		Core   int
		Memory int
	}
	type args struct {
		resource interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "true",
			fields: fields{
				Core:   2,
				Memory: 2,
			},
			args: args{
				resource: &iaas.Server{
					CPU:      2,
					MemoryMB: 4 * size.GiB,
				},
			},
			want: true,
		},
		{
			name: "false",
			fields: fields{
				Core:   2,
				Memory: 4,
			},
			args: args{
				resource: &iaas.Server{
					CPU:      2,
					MemoryMB: 2 * size.GiB,
				},
			},
			want: false,
		},
		{
			name: "false is having same value",
			fields: fields{
				Core:   2,
				Memory: 4,
			},
			args: args{
				resource: &iaas.Server{
					CPU:      2,
					MemoryMB: 4 * size.GiB,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ServerPlan{
				Name:   tt.fields.Name,
				Core:   tt.fields.Core,
				Memory: tt.fields.Memory,
			}
			if got := p.LessThan(tt.args.resource); got != tt.want {
				t.Errorf("LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerPlan_LessThanPlan(t *testing.T) {
	type fields struct {
		Name   string
		Core   int
		Memory int
	}
	type args struct {
		plan ResourcePlan
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "returns true",
			fields: fields{
				Core:   2,
				Memory: 2,
			},
			args: args{
				plan: &ServerPlan{
					Core:   2,
					Memory: 4,
				},
			},
			want: true,
		},
		{
			name: "returns false",
			fields: fields{
				Core:   2,
				Memory: 4,
			},
			args: args{
				plan: &ServerPlan{
					Core:   2,
					Memory: 2,
				},
			},
			want: false,
		},
		{
			name: "returns false if having same value",
			fields: fields{
				Core:   2,
				Memory: 4,
			},
			args: args{
				plan: &ServerPlan{
					Core:   2,
					Memory: 4,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ServerPlan{
				Name:   tt.fields.Name,
				Core:   tt.fields.Core,
				Memory: tt.fields.Memory,
			}
			if got := p.LessThanPlan(tt.args.plan); got != tt.want {
				t.Errorf("LessThanPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}
