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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// Config Coreの起動時に与えられるコンフィギュレーションを保持する
type Config struct {
	SakuraCloud    *SakuraCloud       `yaml:"sakuracloud"`                   // さくらのクラウドAPIのクレデンシャル
	CustomHandlers Handlers           `yaml:"handlers"`                      // カスタムハンドラーの定義
	Resources      *ResourceDefGroups `yaml:"resources" validate:"required"` // リソースグループの定義
	AutoScaler     AutoScalerConfig   `yaml:"autoscaler"`                    // オートスケーラー自体の動作設定
}

// NewConfigFromPath 指定のファイルパスからコンフィギュレーションを読み取ってConfigを作成する
func NewConfigFromPath(filePath string) (*Config, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening configuration file failed: %s error: %s", filePath, err)
	}
	defer reader.Close()

	c := &Config{}
	if err := c.load(reader); err != nil {
		return nil, err
	}

	if c.AutoScaler.ServerTLSConfig != nil {
		c.AutoScaler.ServerTLSConfig.SetDirectory(filepath.Dir(filePath))
	}
	if c.AutoScaler.HandlerTLSConfig != nil {
		c.AutoScaler.HandlerTLSConfig.SetDirectory(filepath.Dir(filePath))
	}
	return c, nil
}

func (c *Config) load(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %s", err)
	}
	if err := yaml.UnmarshalWithOptions(data, c, yaml.Strict()); err != nil {
		return fmt.Errorf(yaml.FormatError(err, false, true))
	}

	if c.SakuraCloud == nil {
		c.SakuraCloud = &SakuraCloud{}
	}
	return nil
}

// APIClient Configに保持しているCredentialからさくらのクラウドAPIクライアント(sacloud.APICaller)を返す
//
// シングルトンなインスタンスを返す
func (c *Config) APIClient() sacloud.APICaller {
	return c.SakuraCloud.APIClient()
}

// Handlers ビルトインハンドラ+カスタムハンドラのリストを返す
//
// ビルトインハンドラはAPIクライアントが注入された状態で返される
func (c *Config) Handlers() Handlers {
	var handlers Handlers
	for _, h := range BuiltinHandlers() {
		if h.Disabled {
			continue
		}
		if h, ok := h.BuiltinHandler.(builtins.SakuraCloudAPICaller); ok {
			h.SetAPICaller(c.SakuraCloud.APIClient())
		}
		handlers = append(handlers, h)
	}
	return append(handlers, c.CustomHandlers...)
}

// Validate 現在のConfig値のバリデーション
func (c *Config) Validate(ctx context.Context) error {
	if err := validate.Struct(c); err != nil {
		return err
	}

	// API Client
	if err := c.SakuraCloud.Validate(ctx); err != nil {
		return err
	}

	// 利用可能ゾーンリストはさくらのクラウドAPIクライアントの設定次第(プロファイルなど)で
	// 変更される可能性があるためこのタイミングで初期化する
	validate.InitValidatorAlias(sacloud.SakuraCloudZones)

	// Resources
	errors := &multierror.Error{}
	if errs := c.Resources.Validate(ctx, c.APIClient(), c.Handlers()); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	// AutoScalerConfig
	if err := c.AutoScaler.Validate(ctx); err != nil {
		errors = multierror.Append(errors, err)
	}

	return errors.ErrorOrNil()
}

// AutoScalerConfig オートスケーラー自体の動作設定
type AutoScalerConfig struct {
	CoolDownSec      int                    `yaml:"cooldown"`           // 同一ジョブの連続実行を防ぐための冷却期間(単位:秒)
	ServerTLSConfig  *config.TLSStruct      `yaml:"server_tls_config"`  // CoreへのgRPC接続のTLS設定
	HandlerTLSConfig *config.TLSStruct      `yaml:"handler_tls_config"` // HandlersへのgRPC接続のTLS設定
	ExporterConfig   *config.ExporterConfig `yaml:"exporter_config"`    // Exporter設定
}

func (c *AutoScalerConfig) Validate(context.Context) error {
	return nil
}

func (c *AutoScalerConfig) JobCoolDownTime() time.Duration {
	sec := c.CoolDownSec
	if sec <= 0 {
		return defaults.CoolDownTime
	}
	return time.Duration(sec) * time.Second
}

func (c *AutoScalerConfig) ExporterEnabled() bool {
	return c.ExporterConfig != nil && c.ExporterConfig.Enabled
}
