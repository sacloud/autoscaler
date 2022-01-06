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

package server

import (
	"reflect"
	"testing"

	"github.com/sacloud/autoscaler/handler"
)

func TestHorizontalScaleHandler_execStartupScriptTemplate(t *testing.T) {
	type args struct {
		server  *handler.ServerGroupInstance
		scripts []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "minimum",
			args: args{
				server:  &handler.ServerGroupInstance{Name: "test"},
				scripts: []string{"{{.Name}}"},
			},
			want:    []string{"test"},
			wantErr: false,
		},
		{
			name: "network interfaces",
			args: args{
				server: &handler.ServerGroupInstance{
					Name: "test",
					NetworkInterfaces: []*handler.ServerGroupInstance_NIC{
						{
							Upstream: "shared",
						},
						{
							Upstream:      "123456789012",
							UserIpAddress: "192.168.11.101",
							AssignedNetwork: &handler.NetworkInfo{
								IpAddress: "192.168.11.101",
								Netmask:   24,
								Gateway:   "192.169.11.1",
								Index:     1,
							},
						},
					},
				},
				scripts: []string{"{{ range .NetworkInterfaces }}{{.Upstream}}{{ if .AssignedNetwork }}:{{ .AssignedNetwork.IpAddress }}{{ end }},{{ end }}"},
			},
			want:    []string{"shared,123456789012:192.168.11.101,"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHorizontalScaleHandler()
			got, err := h.execStartupScriptTemplate(tt.args.server, tt.args.scripts)
			if (err != nil) != tt.wantErr {
				t.Errorf("execStartupScriptTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("execStartupScriptTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
