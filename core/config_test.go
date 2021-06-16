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
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	type fields struct {
		SakuraCloud *SakuraCloud
		Handlers    Handlers
		Resources   *ResourceGroups
		AutoScaler  AutoScalerConfig
	}
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "minimal",
			fields: fields{
				SakuraCloud: &SakuraCloud{
					Credential: Credential{
						Token:  "token",
						Secret: "secret",
					},
				},
				Handlers: Handlers{
					{
						Name:     "fake",
						Endpoint: "unix:autoscaler-handlers-fake.sock",
					},
				},
				Resources: func() *ResourceGroups {
					rgs := newResourceGroups()
					rg := &ResourceGroup{}
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
					}
					rgs.Set("web", rg)
					return rgs
				}(),
				AutoScaler: AutoScalerConfig{
					CoolDownSec: 30,
				},
			},
			args: args{
				reader: bytes.NewReader([]byte(`
sakuracloud:
  token: token
  secret: secret
handlers:
  - name: "fake"
    endpoint: "unix:autoscaler-handlers-fake.sock"
resources:
  web: 
    resources:
      - type: Server
        selector:
          names: ["test-name"]
          zone: "is1a"
        dedicated_cpu: true
autoscaler:
  cooldown: 30
`)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := &Config{
				SakuraCloud:    tt.fields.SakuraCloud,
				CustomHandlers: tt.fields.Handlers,
				Resources:      tt.fields.Resources,
				AutoScaler:     tt.fields.AutoScaler,
			}
			c := &Config{}
			if err := c.load(tt.args.reader); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			require.EqualValues(t, expected, c)
		})
	}
}
