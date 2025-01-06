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

package config

import "github.com/sacloud/autoscaler/defaults"

// ExporterConfig Exporterの設定
type ExporterConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

// ListenAddress Addressが空の場合はデフォルト値(defaults.CoreExporterAddr)を、そうでなければAddressを返す
func (c *ExporterConfig) ListenAddress() string {
	if c.Address == "" {
		return defaults.CoreExporterAddr
	}
	return c.Address
}
