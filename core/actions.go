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
	"context"
	"fmt"
)

// Actions 任意の名称とハンドラー名のリストのマップ
//
// Up/Downリクエスト時にアクション名を指定することで任意のハンドラーだけ実行したい場合に利用する
type Actions map[string][]string

// Validate ハンドラーが実在するか確認する
func (a *Actions) Validate(ctx context.Context, handlers Handlers) []error {
	invalidHandlers := make(map[string]struct{})
	var names []string

	for _, handlerNames := range *a {
		for _, handlerName := range handlerNames {
			exists := false
			for _, h := range handlers {
				if h.Name == handlerName {
					exists = true
					break
				}
			}
			// 同じハンドラ名が複数登場することがあるがエラーは一つにまとめる
			if !exists {
				if _, ok := invalidHandlers[handlerName]; !ok {
					names = append(names, handlerName)
				}
				invalidHandlers[handlerName] = struct{}{}
			}
		}
	}

	var errors []error
	for _, name := range names {
		errors = append(errors, fmt.Errorf("handler %q is not defined", name))
	}
	return errors
}
