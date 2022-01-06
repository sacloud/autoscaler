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

package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sacloud/autoscaler/grpcutil"
	health "google.golang.org/grpc/health/grpc_health_v1"

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
	SakuraCloud    *SakuraCloud        `yaml:"sakuracloud"`                   // さくらのクラウドAPIのクレデンシャル
	CustomHandlers Handlers            `yaml:"handlers"`                      // カスタムハンドラーの定義
	Resources      ResourceDefinitions `yaml:"resources" validate:"required"` // リソースの定義
	AutoScaler     AutoScalerConfig    `yaml:"autoscaler"`                    // オートスケーラー自体の動作設定
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
		if c.handlerDisabled(h) {
			continue
		}
		if h, ok := h.BuiltinHandler.(builtins.SakuraCloudAPICaller); ok {
			h.SetAPICaller(c.SakuraCloud.APIClient())
		}
		handlers = append(handlers, h)
	}
	return append(handlers, c.CustomHandlers...)
}

func (c *Config) handlerDisabled(h *Handler) bool {
	conf := c.AutoScaler.HandlersConfig
	// configで指定されているか?
	if conf != nil {
		// 全体が無効にされているか?
		if conf.Disabled {
			return true
		}
		// 個別に無効にされているか?
		if v, ok := conf.Handlers[h.Name]; ok {
			return v.Disabled
		}
	}
	// デフォルト値
	return h.Disabled
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

	errors := &multierror.Error{}

	if errs := c.ValidateCustomHandlers(ctx); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	// Resources
	if errs := c.Resources.Validate(ctx, c.APIClient()); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	// AutoScalerConfig
	if errs := c.AutoScaler.Validate(ctx); len(errs) > 0 {
		errors = multierror.Append(errors, errs...)
	}

	if len(c.Handlers()) == 0 {
		errors = multierror.Append(errors, fmt.Errorf("one or more handlers are required"))
	}

	return errors.ErrorOrNil()
}

func (c *Config) ValidateCustomHandlers(ctx context.Context) []error {
	var errs []error

	for _, handler := range c.CustomHandlers {
		if err := c.ValidateCustomHandler(ctx, handler); err != nil {
			errs = append(errs, fmt.Errorf("handler %q returns error: %s", handler.Name, err))
		}
	}
	return errs
}

func (c *Config) ValidateCustomHandler(ctx context.Context, handler *Handler) error {
	opt := &grpcutil.DialOption{
		Destination: handler.Endpoint,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("core_to_handlers"),
	}
	if c.AutoScaler.HandlerTLSConfig != nil {
		cred, err := c.AutoScaler.HandlerTLSConfig.TransportCredentials()
		if err != nil {
			return err
		}
		opt.TransportCredentials = cred
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, opt)
	if err != nil {
		return err
	}
	defer cleanup()

	client := health.NewHealthClient(conn)
	res, err := client.Check(ctx, &health.HealthCheckRequest{})
	if err != nil {
		return err
	}
	if res.Status != health.HealthCheckResponse_SERVING {
		return fmt.Errorf("got unexpected status: %s", res.Status)
	}
	return nil
}

// AutoScalerConfig オートスケーラー自体の動作設定
type AutoScalerConfig struct {
	CoolDownSec      int                    `yaml:"cooldown"`           // 同一ジョブの連続実行を防ぐための冷却期間(単位:秒)
	ServerTLSConfig  *config.TLSStruct      `yaml:"server_tls_config"`  // CoreへのgRPC接続のTLS設定
	HandlerTLSConfig *config.TLSStruct      `yaml:"handler_tls_config"` // HandlersへのgRPC接続のTLS設定
	ExporterConfig   *config.ExporterConfig `yaml:"exporter_config"`    // Exporter設定
	HandlersConfig   *HandlersConfig        `yaml:"handlers_config"`    // ビルトインハンドラーの設定
}

func (c *AutoScalerConfig) Validate(ctx context.Context) []error {
	return c.HandlersConfig.Validate(ctx)
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

// HandlersConfig ビルトインハンドラ全体の設定
type HandlersConfig struct {
	Disabled bool                      `yaml:"disabled"` // trueの場合ビルトインハンドラ全体を無効にする
	Handlers map[string]*HandlerConfig `yaml:"handlers"` // ハンドラごとの設定、ハンドラ名をキーにもつ
}

func (c *HandlersConfig) Validate(context.Context) []error {
	if c == nil {
		return nil
	}
	errors := &multierror.Error{}
	for name := range c.Handlers {
		exist := false
		for _, h := range BuiltinHandlers() {
			if h.Name == name {
				exist = true
				break
			}
		}
		if !exist {
			errors = multierror.Append(fmt.Errorf("invalid key: %s", name))
		}
	}
	return errors.Errors
}

// HandlerConfig ビルトインハンドラの設定
type HandlerConfig struct {
	Disabled bool `yaml:"disabled"`
}
