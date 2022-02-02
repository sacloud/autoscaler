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
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/go-homedir"
)

// StringOrFilePath 文字列 or ファイルパス
//
// ファイルパスを指定した場合、ファイルのデータがメモリ内に保持されるため、
// サイズが大きくなるケースでは利用しないようにする
type StringOrFilePath string

func (v *StringOrFilePath) UnmarshalYAML(data []byte) error {
	var str string
	if err := yaml.Unmarshal(data, &str); err != nil {
		return err
	}

	// パスとして存在する場合はファイルを読み取る、そうでない場合はそのまま
	path, err := homedir.Expand(str)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		content = []byte(str)
	}
	*v = StringOrFilePath(content)
	return nil
}

func (v *StringOrFilePath) String() string {
	return string(*v)
}
