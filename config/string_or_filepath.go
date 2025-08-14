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

import (
	"context"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/go-homedir"
	"github.com/sacloud/autoscaler/log"
)

// StringOrFilePath 文字列 or ファイルパス
//
// ファイルパスを指定した場合、ファイルのデータがメモリ内に保持されるため、
// サイズが大きくなるケースでは利用しないようにする
type StringOrFilePath struct {
	content    string
	isFilePath bool
}

func NewStringOrFilePath(ctx context.Context, s string) (*StringOrFilePath, error) {
	strict := false
	logger := log.NewLogger(nil)

	if config, ok := ctx.(LoadConfigHolder); ok {
		strict = config.StrictMode()
	}
	if config, ok := ctx.(LoggerHolder); ok {
		logger = config.Logger()
	}

	if strict {
		return &StringOrFilePath{
			content:    s,
			isFilePath: false,
		}, nil
	}
	content, isFilePath, err := stringOrFilePath(s, logger)
	if err != nil {
		return nil, err
	}
	return &StringOrFilePath{
		content:    content,
		isFilePath: isFilePath,
	}, nil
}

func (v *StringOrFilePath) UnmarshalYAML(ctx context.Context, data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := NewStringOrFilePath(ctx, s)
	if err != nil {
		return err
	}
	*v = *val
	return nil
}

// StringOrFilePath Stringer実装
func (v *StringOrFilePath) String() string {
	return v.content
}

// Bytes .
func (v *StringOrFilePath) Bytes() []byte {
	return []byte(v.content)
}

// Empty vの文字列、またはvがファイルパスの場合はファイルの内容が空だった場合にtrueを返す
func (v *StringOrFilePath) Empty() bool {
	return v.content == ""
}

// IsFilePath vの文字列がファイルパスであるかの判定結果を返す
func (v *StringOrFilePath) IsFilePath() bool {
	return v.isFilePath
}

func stringOrFilePath(s string, logger *slog.Logger) (string, bool, error) {
	if logger == nil {
		logger = log.NewLogger(nil)
	}

	path, err := homedir.Expand(s)
	if err != nil {
		return "", false, err
	}

	isFilePath := true
	content, err := os.ReadFile(path) //nolint:gosec

	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warn(
				"got unknown error while processing StringOrFilePath",
				slog.Any("error", err),
			)
		}
		isFilePath = false
		content = []byte(s)
	}
	return string(content), isFilePath, nil
}
