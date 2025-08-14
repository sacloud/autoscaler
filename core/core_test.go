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
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/test"
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
			name: "default resource name with nested definition",
			resources: ResourceDefinitions{
				&stubResourceDef{
					ResourceDefBase: &ResourceDefBase{
						TypeName: "stub",
						DefName:  "name1",
					},
				},
			},
			args:    defaults.ResourceName,
			want:    "name1",
			wantErr: false,
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

func TestLoadAndValidate(t *testing.T) {
	os.Setenv("SAKURACLOUD_FAKE_MODE", "1") //nolint:errcheck
	defer test.AddTestELB(t, "example")()

	type args struct {
		configPath string
		strictMode bool
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				configPath: "",
				strictMode: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "minimum",
			args: args{
				configPath: "./test/minimum.yaml",
				strictMode: false,
			},
			want: &Config{
				SakuraCloud: &SakuraCloud{},
				Resources: ResourceDefinitions{
					&ResourceDefELB{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "EnhancedLoadBalancer",
							DefName:  "example",
						},
						Selector: &ResourceSelector{
							Names: []string{"example"},
						},
					},
				},
				AutoScaler: AutoScalerConfig{
					CoolDown: &CoolDown{
						Up:   5,
						Down: 5,
					},
				},
				strictMode: false,
			},
			wantErr: false,
		},
		{
			name: "minimum with strict mode",
			args: args{
				configPath: "./test/minimum.yaml",
				strictMode: true,
			},
			want: &Config{
				SakuraCloud: &SakuraCloud{},
				Resources: ResourceDefinitions{
					&ResourceDefELB{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "EnhancedLoadBalancer",
							DefName:  "example",
						},
						Selector: &ResourceSelector{
							Names: []string{"example"},
						},
					},
				},
				AutoScaler: AutoScalerConfig{
					CoolDown: &CoolDown{
						Up:   5,
						Down: 5,
					},
				},
				strictMode: true, // 引数での指定がConfigに引き継がれているはず
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadAndValidate(context.Background(), tt.args.configPath, tt.args.strictMode, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				require.Equal(t, tt.want.AutoScaler, got.AutoScaler)
				require.Equal(t, tt.want.CustomHandlers, got.CustomHandlers)

				require.Equal(t, tt.want.SakuraCloud.Credential, got.SakuraCloud.Credential)
				require.Equal(t, tt.want.SakuraCloud.Profile, got.SakuraCloud.Profile)

				require.Equal(t, tt.want.Resources, got.Resources)
				require.Equal(t, tt.want.strictMode, got.strictMode)
			}
		})
	}
}

func TestCore_Stop(t *testing.T) {
	t.Run("with not running state", func(t *testing.T) {
		c := &Core{
			logger: log.NewLogger(&log.LoggerOption{
				Writer: os.Stderr,
				Level:  slog.LevelInfo,
			}),
		}

		err := c.stop(5 * time.Second)
		require.NoError(t, err)
		require.True(t, c.stopping)
	})

	t.Run("with running state", func(t *testing.T) {
		c := &Core{
			logger: log.NewLogger(&log.LoggerOption{
				Writer: os.Stderr,
				Level:  slog.LevelInfo,
			}),
			running: true,
		}

		// 一定時間後にrunningをfalseへ
		go func() {
			time.Sleep(1 * time.Second)
			c.setRunningStatus(false)
		}()
		err := c.stop(10 * time.Second)
		require.NoError(t, err)
	})

	t.Run("with running state and expect timeout", func(t *testing.T) {
		c := &Core{
			logger: log.NewLogger(&log.LoggerOption{
				Writer: os.Stderr,
				Level:  slog.LevelInfo,
			}),
			running: true,
		}

		err := c.stop(time.Millisecond)
		require.Error(t, err) // timeout
	})
}
