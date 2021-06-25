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

package inputs

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/config"
)

// TLSConfig InputsのTLS関連動作設定
type TLSConfig struct {
	// WebServer WebhookサーバのTLS関連設定
	Server *config.TLSStruct `yaml:"server"`
	// CoreClient coreのgRPCクライアントのTLS関連設定
	CoreClient *config.TLSStruct `yaml:"core_client"`
}

// LoadTLSConfigFromPath 指定のパスからConfigをロードする
func LoadTLSConfigFromPath(path string) (*TLSConfig, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &TLSConfig{}
	if err := yaml.UnmarshalWithOptions(data, c, yaml.Strict()); err != nil {
		return nil, err
	}

	if c.Server != nil {
		c.Server.SetDirectory(filepath.Dir(path))
	}
	if c.CoreClient != nil {
		c.CoreClient.SetDirectory(filepath.Dir(path))
	}

	return c, nil
}
