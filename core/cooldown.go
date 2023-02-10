// Copyright 2021-2023 The sacloud/autoscaler Authors
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

	"github.com/goccy/go-yaml"
)

// CoolDown スケール動作を繰り返し実行する際の冷却期間
type CoolDown struct {
	Up   int `yaml:"up"`
	Down int `yaml:"down"`
}

func (c *CoolDown) UnmarshalYAML(ctx context.Context, data []byte) error {
	// まずintで指定されているか確認
	var cd int
	if err := yaml.UnmarshalContext(ctx, data, &cd); err == nil { // エラーなくUnmarshalできたら
		*c = CoolDown{
			Up:   cd,
			Down: cd,
		}
		return nil
	}

	// int以外の場合はstructとしてUnmarshal
	type alias CoolDown
	var v alias
	if err := yaml.UnmarshalContext(ctx, data, &v); err != nil {
		return err
	}
	*c = CoolDown(v)
	return nil
}
