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
	"os"

	"github.com/goccy/go-yaml"
)

// NameOrSelector 名前(文字列)、もしくはResourceSelectorを表すstruct
type NameOrSelector struct {
	ResourceSelector
}

func (v *NameOrSelector) UnmarshalYAML(data []byte) error {
	// セレクタとしてUnmarshalしてみてエラーだったら文字列と見なす
	var selector ResourceSelector
	if err := yaml.UnmarshalWithOptions(data, &selector, yaml.Strict()); err != nil {
		selector = ResourceSelector{
			Names: []string{string(data)},
		}
	}
	*v = NameOrSelector{ResourceSelector: selector}
	return nil
}

// StringOrFilePath 文字列 or ファイルパス
type StringOrFilePath string

func (v *StringOrFilePath) UnmarshalYAML(data []byte) error {
	// パスとして存在する場合はファイルを読み取る、そうでない場合はそのまま
	content, err := os.ReadFile(string(data))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		content = data
	}
	*v = StringOrFilePath(content)
	return nil
}
