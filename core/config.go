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
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// Config Coreの起動時に与えられるコンフィギュレーションを保持する
type Config struct {
	SakuraCloud *SakuraCloud   `yaml:"sakuracloud"` // さくらのクラウドAPIのクレデンシャル
	Actions     Actions        `yaml:"actions"`     // Inputsからのリクエストパラメータとして指定されるアクションリストのマップ、Inputsからはキーを指定する
	Handlers    Handlers       `yaml:"handlers"`    // カスタムハンドラーの定義
	Resources   ResourceGroups `yaml:"resources"`   // リソースグループの定義
}

func NewConfigFromPath(filePath string) (*Config, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening configuration file failed: %s error: %s", filePath, err)
	}
	defer reader.Close()

	return NewConfigFromReader(reader)
}

func NewConfigFromReader(reader io.Reader) (*Config, error) {
	c := &Config{}
	if err := c.load(reader); err != nil {
		return nil, err
	}
	return c, nil
}

// Load 指定のreaderからYAMLを読み取りConfigへ値を設定する
func (c *Config) load(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %s", err)
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshalling of config values failed: %s", err)
	}
	return nil
}

// APIClient Configに保持しているCredentialからさくらのクラウドAPIクライアント(sacloud.APICaller)を返す
func (c *Config) APIClient() sacloud.APICaller {
	return api.NewCaller(&api.CallerOptions{
		AccessToken:       c.SakuraCloud.Token,
		AccessTokenSecret: c.SakuraCloud.Secret,
		//APIRootURL:           "",
		//DefaultZone:          "",
		//AcceptLanguage:       "",
		//HTTPClient:           nil,
		//HTTPRequestTimeout:   0,
		//HTTPRequestRateLimit: 0,
		//RetryMax:             0,
		//RetryWaitMax:         0,
		//RetryWaitMin:         0,
		//UserAgent:            "",
		//TraceAPI:             false,
		//TraceHTTP:            false,
		//OpenTelemetry:        false,
		//OpenTelemetryOptions: nil,
		FakeMode: true, // TODO プロトタイプはFakeモードで動作させる
		//FakeStorePath:        "",
	})
}
