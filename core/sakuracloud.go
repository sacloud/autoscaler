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
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/go-otelsetup"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/defaults"
	"github.com/sacloud/iaas-api-go/helper/api"
	sacloudtrace "github.com/sacloud/iaas-api-go/trace/otel"
)

type SakuraCloud struct {
	Credential `yaml:",inline"`
	Profile    string `yaml:"profile"`

	strictMode bool
	apiClient  iaas.APICaller
	initOnce   sync.Once
	initError  error
}

// APIClient シングルトンなAPIクライアントを返す
func (sc *SakuraCloud) APIClient() iaas.APICaller {
	sc.initOnce.Do(func() {
		options := []*api.CallerOptions{
			api.OptionsFromEnv(),
		}
		if !sc.strictMode {
			opt, err := api.OptionsFromProfile(sc.Profile)
			if err != nil {
				sc.initError = err
				return
			}
			options = append(options, opt)
		}

		options = append(options, &api.CallerOptions{
			Options: &client.Options{
				AccessToken:       sc.Token,
				AccessTokenSecret: sc.Secret,
				UserAgent: fmt.Sprintf(
					"sacloud/autoscaler/v%s (%s/%s; +https://github.com/sacloud/autoscaler) %s",
					version.Version,
					runtime.GOOS,
					runtime.GOARCH,
					os.Getenv("SAKURACLOUD_APPEND_USER_AGENT"),
				),
			},
		})
		sc.apiClient = api.NewCallerWithOptions(api.MergeOptions(options...))
		// オートスケールでは時間のかかる状態変更待ち(大きなディスクのコピー待ちなど)はあまりない想定
		defaults.DefaultStatePollingTimeout = 60 * time.Minute

		// 環境変数OTEL_EXPORTER_OTLP_ENDPOINTが指定されていたらOpenTelemetryによるトレースを有効化
		if otelsetup.Enabled() {
			sacloudtrace.Initialize()
		}
	})
	return sc.apiClient
}

// Validate 有効なAPIキーが指定されているかを確認する
func (sc *SakuraCloud) Validate(ctx context.Context) error {
	if len(os.Getenv("SAKURACLOUD_APPEND_USER_AGENT")) > 1024 {
		return fmt.Errorf("SAKURACLOUD_APPEND_USER_AGENT is too long: max=1024")
	}

	apiClient := sc.APIClient()
	if sc.initError != nil {
		return fmt.Errorf("initializing API Client failed: %s", sc.initError)
	}

	// APIクライアントが有効かどうかを確認するためにゾーンのリストを取得してみる
	_, err := iaas.NewZoneOp(apiClient).Find(ctx, nil)
	if err != nil {
		var apiErr iaas.APIError
		if errors.As(err, &apiErr) {
			return validate.Errorf("failed to call SAKURA Cloud API: %s", apiErr.Message())
		}
		return fmt.Errorf("failed to call SAKURA Cloud API: %w", err)
	}
	return nil
}
