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
	"time"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// Config Coreの起動時に与えられるコンフィギュレーションを保持する
type Config struct {
	SakuraCloud    *SakuraCloud       `yaml:"sakuracloud"` // さくらのクラウドAPIのクレデンシャル
	CustomHandlers Handlers           `yaml:"handlers"`    // カスタムハンドラーの定義
	Resources      *ResourceDefGroups `yaml:"resources"`   // リソースグループの定義
	AutoScaler     AutoScalerConfig   `yaml:"autoscaler"`  // オートスケーラー自体の動作設定
}

// NewConfigFromPath 指定のファイルパスからコンフィギュレーションを読み取ってConfigを作成する
func NewConfigFromPath(filePath string) (*Config, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening configuration file failed: %s error: %s", filePath, err)
	}
	defer reader.Close()

	return NewConfigFromReader(reader)
}

// NewConfigFromReader 指定のio.Readerからコンフィギュレーションを読み取ってConfigを作成する
func NewConfigFromReader(reader io.Reader) (*Config, error) {
	c := &Config{}
	if err := c.load(reader); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) load(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %s", err)
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshalling of config values failed: %s", err)
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
	handlers := BuiltinHandlers()
	for _, h := range handlers {
		if h, ok := h.BuiltinHandler.(builtins.SakuraCloudAPICaller); ok {
			h.SetAPICaller(c.SakuraCloud.APIClient())
		}
	}
	return append(handlers, c.CustomHandlers...)
}

// Validate 現在のConfig値のバリデーション
func (c *Config) Validate(ctx context.Context) error {
	// API Client
	if err := c.SakuraCloud.Validate(ctx); err != nil {
		return err
	}

	// Resources
	errors := &multierror.Error{}
	if errs := c.Resources.Validate(ctx, c.APIClient(), c.Handlers()); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	return errors.ErrorOrNil()
}

// AutoScalerConfig オートスケーラー自体の動作設定
type AutoScalerConfig struct {
	CoolDownSec int `yaml:"cooldown"` // 同一ジョブの連続実行を防ぐための冷却期間(単位:秒)
}

func (c *AutoScalerConfig) JobCoolDownTime() time.Duration {
	sec := c.CoolDownSec
	if sec <= 0 {
		return defaults.CoolDownTime
	}
	return time.Duration(sec) * time.Second
}
