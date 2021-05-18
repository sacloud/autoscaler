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

	"github.com/goccy/go-yaml"
)

// Config Coreの起動時に与えられるコンフィギュレーションを保持する
type Config struct {
	SakuraCloud *Credential    `json:"sakuracloud" yaml:"sakuracloud"` // さくらのクラウドAPIのクレデンシャル
	Actions     Actions        `json:"actions" yaml:"actions"`         // Inputsからのリクエストパラメータとして指定されるアクションリストのマップ、Inputsからはキーを指定する
	Handlers    Handlers       `json:"handlers" yaml:"handlers"`       // カスタムハンドラーの定義
	Resources   ResourceGroups `json:"resources" yaml:"resources"`     // リソースグループの定義
}

// Load 指定のreaderからYAMLを読み取りConfigへ値を設定する
func (c *Config) Load(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %s", err)
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshalling of config values failed: %s", err)
	}
	return nil
}
