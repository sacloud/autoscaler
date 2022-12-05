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

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/iaas-api-go/types"
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
				HttpClient:        &http.Client{},
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
	})
	return sc.apiClient
}

// Validate 有効なAPIキーが指定されており、必要なパーミッションが割り当てられていることを確認する
func (sc *SakuraCloud) Validate(ctx context.Context) error {
	apiClient := sc.APIClient()
	if sc.initError != nil {
		return fmt.Errorf("initializing API Client failed: %s", sc.initError)
	}

	authStatus, err := iaas.NewAuthStatusOp(apiClient).Read(ctx)
	if err != nil {
		if err, ok := err.(iaas.APIError); ok {
			return validate.Errorf("reading SAKURA cloud account info failed: %s", err.Message())
		}
		return fmt.Errorf("reading SAKURA cloud account info failed: unknown error: %s", err)
	}
	if authStatus.Permission != types.Permissions.Create && authStatus.Permission != types.Permissions.Arrange {
		return validate.Errorf("required permissions have not been assigned. assigned permission: %s", authStatus.Permission)
	}
	if len(os.Getenv("SAKURACLOUD_APPEND_USER_AGENT")) > 1024 {
		return fmt.Errorf("SAKURACLOUD_APPEND_USER_AGENT is too long: max=1024")
	}
	return nil
}
