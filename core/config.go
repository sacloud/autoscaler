// Copyright 2021-2023 The sacloud/autoscaler Authors
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
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

// Config Coreの起動時に与えられるコンフィギュレーションを保持する
type Config struct {
	SakuraCloud    *SakuraCloud        `yaml:"sakuracloud"`                   // さくらのクラウドAPIのクレデンシャル
	CustomHandlers Handlers            `yaml:"handlers"`                      // カスタムハンドラーの定義
	Resources      ResourceDefinitions `yaml:"resources" validate:"required"` // リソースの定義
	AutoScaler     AutoScalerConfig    `yaml:"autoscaler"`                    // オートスケーラー自体の動作設定

	strictMode bool
	logger     *log.Logger
}

// NewConfigFromPath 指定のファイルパスからコンフィギュレーションを読み取ってConfigを作成する
func NewConfigFromPath(ctx context.Context, filePath string, strictMode bool, logger *log.Logger) (*Config, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening configuration file failed: %s error: %s", filePath, err)
	}
	defer reader.Close()

	c := &Config{strictMode: strictMode, logger: logger}
	if err := c.load(ctx, reader); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) load(ctx context.Context, reader io.Reader) error {
	ctx = config.NewLoadConfigContext(ctx, c.strictMode, c.logger)

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %s", err)
	}
	if err := yaml.UnmarshalContext(ctx, data, c, yaml.Strict()); err != nil {
		return fmt.Errorf(yaml.FormatError(err, false, true))
	}

	if c.SakuraCloud == nil {
		c.SakuraCloud = &SakuraCloud{}
	}
	c.SakuraCloud.strictMode = c.strictMode
	return nil
}

// APIClient Configに保持しているCredentialからさくらのクラウドAPIクライアント(iaas.APICaller)を返す
//
// シングルトンなインスタンスを返す
func (c *Config) APIClient() iaas.APICaller {
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
	ctx = config.NewLoadConfigContext(ctx, c.strictMode, c.logger)

	if err := validate.Struct(c); err != nil {
		return validate.New(err)
	}

	// API Client(ValidationErrorを返すことがある)
	if err := c.SakuraCloud.Validate(ctx); err != nil {
		return err
	}

	// 利用可能ゾーンリストはさくらのクラウドAPIクライアントの設定次第(プロファイルなど)で
	// 変更される可能性があるためこのタイミングで初期化する
	validate.InitValidatorAlias(iaas.SakuraCloudZones)

	allErrors := &multierror.Error{}

	// CustomHandlers
	if errs := c.ValidateCustomHandlers(ctx); len(errs) > 0 {
		allErrors = multierror.Append(allErrors, errs...)
	}

	// Resources
	if errs := c.Resources.Validate(ctx, c.APIClient()); len(errs) > 0 {
		allErrors = multierror.Append(allErrors, errs...)
	}

	// AutoScalerConfig
	if errs := c.AutoScaler.Validate(ctx); len(errs) > 0 {
		allErrors = multierror.Append(allErrors, errs...)
	}

	// All Handlers (Builtin + Custom)
	if len(c.Handlers()) == 0 {
		allErrors = multierror.Append(allErrors, validate.Errorf("one or more handlers are required"))
	}

	if c.strictMode {
		// プロファイル指定を制限
		if c.SakuraCloud.Profile != "" {
			allErrors = multierror.Append(allErrors, validate.Errorf("sakuracloud.profile cannot be specified when in strict mode"))
		}
		// exporterを有効にすることを制限
		if c.AutoScaler.ExporterEnabled() {
			allErrors = multierror.Append(allErrors, validate.Errorf("autoscaler.exporter_config cannot be specified when in strict mode"))
		}
		// カスタムハンドラを定義することを制限
		if len(c.CustomHandlers) > 0 {
			allErrors = multierror.Append(allErrors, validate.Errorf("handlers cannot be specified when in strict mode"))
		}
	}

	hasSystemError := false
	for _, err := range allErrors.Errors {
		if !errwrap.ContainsType(err, &validate.Error{}) {
			hasSystemError = true
			break
		}
	}
	if hasSystemError {
		return allErrors.ErrorOrNil()
	}

	return validate.New(allErrors.ErrorOrNil())
}

func (c *Config) ValidateCustomHandlers(ctx context.Context) []error {
	var errs []error

	for _, handler := range c.CustomHandlers {
		if err := c.ValidateCustomHandler(ctx, handler); err != nil {
			errs = append(errs, validate.Errorf("handler %q returns error: %s", handler.Name, err))
		}
	}
	return errs
}

func (c *Config) ValidateCustomHandler(ctx context.Context, handler *Handler) error {
	opt := &grpcutil.DialOption{
		Destination: handler.Endpoint,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("core_to_handlers"),
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
	CoolDownSec            int                    `yaml:"cooldown"`              // 同一ジョブの連続実行を防ぐための冷却期間(単位:秒)
	ShutdownGracePeriodSec int                    `yaml:"shutdown_grace_period"` // SIGINTまたはSIGTERMをを受け取った際の処理完了待ち猶予時間(単位:秒)
	ExporterConfig         *config.ExporterConfig `yaml:"exporter_config"`       // Exporter設定
	HandlersConfig         *HandlersConfig        `yaml:"handlers_config"`       // ビルトインハンドラーの設定
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

func (c *AutoScalerConfig) ShutdownGracePeriod() time.Duration {
	sec := c.ShutdownGracePeriodSec
	if sec <= 0 {
		return defaults.ShutdownGracePeriod
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
			errors = multierror.Append(validate.Errorf("invalid key: %s", name))
		}
	}
	return errors.Errors
}

// HandlerConfig ビルトインハンドラの設定
type HandlerConfig struct {
	Disabled bool `yaml:"disabled"`
}
