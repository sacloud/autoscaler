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
	"strings"
	"sync"
	"time"

	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/defaults"
	"github.com/sacloud/sacloud-sdk-go/common/packages/envvar"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
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
		var sa saclient.Client

		env := os.Environ()
		if sc.Token != "" && sc.Secret != "" {
			env = append(env, "SAKURA_ACCESS_TOKEN="+sc.Token)
			env = append(env, "SAKURA_ACCESS_TOKEN_SECRET="+sc.Secret)
		}
		if sc.Profile != "" {
			env = append(env, "SAKURA_PROFILE="+sc.Profile)
		}
		if err := sa.SetEnviron(env); err != nil {
			sc.initError = err
			return
		}

		appendUA := envvar.StringFromEnvMulti([]string{"SAKURA_APPEND_USER_AGENT", "SAKURACLOUD_APPEND_USER_AGENT"}, "")
		ua := fmt.Sprintf(
			"sacloud/autoscaler/v%s (%s/%s; +https://github.com/sacloud/autoscaler) %s",
			version.Version,
			runtime.GOOS,
			runtime.GOARCH,
			appendUA,
		)

		var err error

		if sc.strictMode {
			err = sa.SetWith(
				saclient.WithUserAgent(ua),
				saclient.WithoutProfile(),
			)
		} else {
			err = sa.SetWith(saclient.WithUserAgent(ua))
		}
		if err != nil {
			sc.initError = err
			return
		}

		if err := sa.Populate(); err != nil {
			sc.initError = err
			return
		}

		cfg, err := sa.EndpointConfig()
		if err != nil {
			sc.initError = fmt.Errorf("failed to get endpoint config: %w", err)
			return
		}
		if cfg.APIRootURL != "" {
			if strings.HasSuffix(cfg.APIRootURL, "/") {
				cfg.APIRootURL = strings.TrimRight(cfg.APIRootURL, "/")
			}
			iaas.SakuraCloudAPIRoot = cfg.APIRootURL
		}
		if len(cfg.Zones) > 0 {
			iaas.SakuraCloudZones = cfg.Zones
		}
		if cfg.DefaultZone != "" {
			iaas.APIDefaultZone = cfg.DefaultZone
		}

		sc.apiClient = iaas.NewClientFromSaclient(&sa)
		// オートスケールでは時間のかかる状態変更待ち(大きなディスクのコピー待ちなど)はあまりない想定
		defaults.DefaultStatePollingTimeout = 60 * time.Minute
	})
	return sc.apiClient
}

// Validate 有効なAPIキーが指定されているかを確認する
func (sc *SakuraCloud) Validate(ctx context.Context) error {
	appendUA := envvar.StringFromEnvMulti([]string{"SAKURA_APPEND_USER_AGENT", "SAKURACLOUD_APPEND_USER_AGENT"}, "")
	if len(appendUA) > 1024 {
		return fmt.Errorf("SAKURA_APPEND_USER_AGENT (or SAKURACLOUD_APPEND_USER_AGENT) is too long: max=1024")
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
