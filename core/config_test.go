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
	"context"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

func TestHandlersConfig_UnmarshalYAML(t *testing.T) {
	data := []byte(`
disabled: true
handlers:
  foo:
    disabled: true
  dns-servers-handler:
    disabled: true
`)

	var config HandlersConfig
	if err := yaml.UnmarshalWithOptions(data, &config); err != nil {
		t.Fatal(err)
	}
	expected := HandlersConfig{
		Disabled: true,
		Handlers: map[string]*HandlerConfig{
			"foo":                 {Disabled: true},
			"dns-servers-handler": {Disabled: true},
		},
	}
	require.EqualValues(t, expected, config)
}

func TestConfig_Handlers(t *testing.T) {
	type fields struct {
		AutoScaler AutoScalerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   Handlers
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   BuiltinHandlers(),
		},
		{
			name: "disable all",
			fields: fields{
				AutoScaler: AutoScalerConfig{
					HandlersConfig: &HandlersConfig{
						Disabled: true,
					},
				},
			},
			want: nil,
		},
		{
			name: "disable per handler",
			fields: fields{
				AutoScaler: AutoScalerConfig{
					HandlersConfig: &HandlersConfig{
						Handlers: map[string]*HandlerConfig{
							"dns-servers-handler":           {Disabled: false},
							"elb-vertical-scaler":           {Disabled: false},
							"elb-servers-handler":           {Disabled: true},
							"gslb-servers-handler":          {Disabled: true},
							"load-balancer-servers-handler": {Disabled: true},
							"router-vertical-scaler":        {Disabled: true},
							"server-horizontal-scaler":      {Disabled: true},
							"server-vertical-scaler":        {Disabled: true},
						},
					},
				},
			},
			want: []*Handler{
				{Name: "dns-servers-handler"},
				{Name: "elb-vertical-scaler"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				SakuraCloud: &SakuraCloud{Credential: Credential{}},
				AutoScaler:  tt.fields.AutoScaler,
			}
			got := c.Handlers()

			var gotNames, wantNames []string
			for _, h := range got {
				gotNames = append(gotNames, h.Name)
			}
			for _, h := range tt.want {
				wantNames = append(wantNames, h.Name)
			}
			require.EqualValues(t, gotNames, wantNames)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	os.Setenv("SAKURACLOUD_FAKE_MODE", "1") //nolint:errcheck
	defer test.AddTestELB(t, "example")()

	resources := ResourceDefinitions{
		&ResourceDefELB{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "EnhancedLoadBalancer",
				DefName:  "example",
			},
			Selector: &ResourceSelector{
				Names: []string{"example"},
			},
		},
	}

	type fields struct {
		SakuraCloud    *SakuraCloud
		CustomHandlers Handlers
		Resources      ResourceDefinitions
		AutoScaler     AutoScalerConfig
		strictMode     bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "minimum",
			fields: fields{
				SakuraCloud: &SakuraCloud{strictMode: false},
				Resources:   resources,
			},
			wantErr: false,
		},
		{
			name: "strict with sakuracloud.profile",
			fields: fields{
				strictMode:  true,
				SakuraCloud: &SakuraCloud{strictMode: true, Profile: "foobar"},
				Resources:   resources,
			},
			wantErr: true,
		},
		{
			name: "strict with exporter",
			fields: fields{
				strictMode:  true,
				SakuraCloud: &SakuraCloud{strictMode: true},
				Resources:   resources,
				AutoScaler: AutoScalerConfig{
					ExporterConfig: &config.ExporterConfig{
						Enabled: true,
						Address: ":8080",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "strict with custom handlers",
			fields: fields{
				strictMode:  true,
				SakuraCloud: &SakuraCloud{strictMode: true},
				Resources:   resources,
				CustomHandlers: Handlers{
					{
						Name:     "example",
						Endpoint: "unix:example.sock",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				SakuraCloud:    tt.fields.SakuraCloud,
				CustomHandlers: tt.fields.CustomHandlers,
				Resources:      tt.fields.Resources,
				AutoScaler:     tt.fields.AutoScaler,
				strictMode:     tt.fields.strictMode,
				logger:         log.NewLogger(nil),
			}
			if err := c.Validate(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
