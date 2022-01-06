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
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2"
	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type SakuraCloud struct {
	Credential `yaml:",inline"`
	// TODO 項目追加

	apiClient sacloud.APICaller
	initOnce  sync.Once
}

// APIClient シングルトンなAPIクライアントを返す
func (sc *SakuraCloud) APIClient() sacloud.APICaller {
	sc.initOnce.Do(func() {
		token := sc.Token
		if token == "" {
			token = os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
		}
		secret := sc.Secret
		if secret == "" {
			secret = os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
		}

		ua := fmt.Sprintf(
			"sacloud/autoscaler/v%s (%s/%s; +https://github.com/sacloud/autoscaler) libsacloud/%s",
			version.Version,
			runtime.GOOS,
			runtime.GOARCH,
			libsacloud.Version,
		)

		httpClient := &http.Client{}
		sc.apiClient = api.NewCaller(&api.CallerOptions{
			AccessToken:          token,
			AccessTokenSecret:    secret,
			HTTPClient:           httpClient,
			HTTPRequestTimeout:   300,
			HTTPRequestRateLimit: 10,
			RetryMax:             10,
			UserAgent:            ua,
			TraceAPI:             os.Getenv("SAKURACLOUD_TRACE") != "",
			TraceHTTP:            os.Getenv("SAKURACLOUD_TRACE") != "",
			FakeMode:             os.Getenv("FAKE_MODE") != "",
		})
	})
	return sc.apiClient
}

// Validate 有効なAPIキーが指定されており、必要なパーミッションが割り当てられていることを確認する
func (sc *SakuraCloud) Validate(ctx context.Context) error {
	authStatus, err := sacloud.NewAuthStatusOp(sc.APIClient()).Read(ctx)
	if err != nil {
		if err, ok := err.(sacloud.APIError); ok {
			return fmt.Errorf("reading SAKURA cloud account info failed: %s", err.Message())
		}
		return fmt.Errorf("reading SAKURA cloud account info failed: unknown error: %s", err)
	}
	if authStatus.Permission != types.Permissions.Create && authStatus.Permission != types.Permissions.Arrange {
		return fmt.Errorf("required permissions have not been assigned. assigned permission: %s", authStatus.Permission)
	}
	return nil
}
