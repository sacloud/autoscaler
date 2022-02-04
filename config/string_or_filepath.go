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

package config

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/go-homedir"
)

// StringOrFilePath 文字列 or ファイルパス
//
// ファイルパスを指定した場合、ファイルのデータがメモリ内に保持されるため、
// サイズが大きくなるケースでは利用しないようにする
type StringOrFilePath struct {
	content    string
	isFilePath bool
}

func NewStringOrFilePath(s string) (*StringOrFilePath, error) {
	content, isFilePath, err := stringOrFilePath(s)
	if err != nil {
		return nil, err
	}
	return &StringOrFilePath{
		content:    content,
		isFilePath: isFilePath,
	}, nil
}

func (v *StringOrFilePath) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := NewStringOrFilePath(s)
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

func stringOrFilePath(s string) (string, bool, error) {
	path, err := homedir.Expand(s)
	if err != nil {
		return "", false, err
	}

	isFilePath := true
	content, err := os.ReadFile(path)

	if err != nil {
		// Note:
		// ファイル不存在以外のエラーも全て無視している。
		// このためエラーが発生していても警告を出さずに処理を進めてしまう。
		// 運用上問題になるケースは少ないと思われるが、将来的にここでログ出力が行えるようになったら対応すべき。
		isFilePath = false
		content = []byte(s)
	}
	return string(content), isFilePath, nil
}
