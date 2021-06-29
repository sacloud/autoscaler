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

	"github.com/sacloud/autoscaler/config"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	type fields struct {
		SakuraCloud *SakuraCloud
		Handlers    Handlers
		Resources   *ResourceDefGroups
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
				Resources: func() *ResourceDefGroups {
					rgs := newResourceDefGroups()
					rg := &ResourceDefGroup{}
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
					}
					rgs.Set("web", rg)
					return rgs
				}(),
				AutoScaler: AutoScalerConfig{
					CoolDownSec: 30,
					ServerTLSConfig: &config.TLSStruct{
						TLSCertPath: "server.crt",
						TLSKeyPath:  "server.key",
						ClientAuth:  "RequireAndVerifyClientCert",
						ClientCAs:   "ca.crt",
					},
					HandlerTLSConfig: &config.TLSStruct{
						TLSCertPath: "server.crt",
						TLSKeyPath:  "server.key",
						RootCAs:     "ca.crt",
					},
					ExporterConfig: &config.ExporterConfig{
						Enabled: true,
						Address: "localhost:8080",
						TLSConfig: &config.TLSStruct{
							TLSCertPath: "server.crt",
							TLSKeyPath:  "server.key",
							ClientAuth:  "RequireAndVerifyClientCert",
							ClientCAs:   "ca.crt",
						},
					},
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
          zones: ["is1a"]
        dedicated_cpu: true
autoscaler:
  cooldown: 30
  server_tls_config:
    cert_file: server.crt
    key_file: server.key
    client_auth_type: RequireAndVerifyClientCert
    client_ca_file: ca.crt
  handler_tls_config:
    cert_file: server.crt
    key_file: server.key
    root_ca_file: ca.crt
  exporter_config:
    enabled: true
    address: "localhost:8080"
    tls_config:
      cert_file: server.crt
      key_file: server.key
      client_auth_type: RequireAndVerifyClientCert
      client_ca_file: ca.crt

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
