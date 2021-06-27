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

// Config InputsのTLS関連動作設定
type Config struct {
	// ServerTLSConfig WebhookサーバのTLS関連設定
	ServerTLSConfig *config.TLSStruct `yaml:"server_tls_config"`

	// CoreTLSConfig coreのgRPCクライアントのTLS関連設定
	CoreTLSConfig *config.TLSStruct `yaml:"core_tls_config"`

	// ExporterConfig Exporterの設定
	ExporterConfig *config.ExporterConfig `yaml:"exporter_config"`
}

// LoadConfigFromPath 指定のパスからConfigをロードする
func LoadConfigFromPath(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := yaml.UnmarshalWithOptions(data, c, yaml.Strict()); err != nil {
		return nil, err
	}

	if c.ServerTLSConfig != nil {
		c.ServerTLSConfig.SetDirectory(filepath.Dir(path))
	}
	if c.CoreTLSConfig != nil {
		c.CoreTLSConfig.SetDirectory(filepath.Dir(path))
	}
	if c.ExporterConfig != nil && c.ExporterConfig.TLSConfig != nil {
		c.ExporterConfig.TLSConfig.SetDirectory(filepath.Dir(path))
	}

	return c, nil
}
