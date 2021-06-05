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

// HandlingContext 1リクエスト中の1リソースに対するハンドリングのスコープに対応するコンテキスト
//
// context.Contextを実装し、core.Contextに加えて現在処理中の[]Computedを保持する
type HandlingContext struct {
	*Context
	currentComputed []Computed
}

func NewHandlingContext(parent *Context, computed []Computed) *HandlingContext {
	return &HandlingContext{
		Context:         parent,
		currentComputed: computed,
	}
}

// CurrentComputed 現在処理中の[]Computedを返す
func (c *HandlingContext) CurrentComputed() []Computed {
	return c.currentComputed
}
