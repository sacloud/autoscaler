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

type Handlers []*Handler

// Handler カスタムハンドラーの定義
type Handler struct {
	Type     string `json:"type" yaml:"type"` // ハンドラー種別 TODO: enumにすべきか要検討
	Name     string `json:"name" yaml:"name"` // ハンドラーを識別するための名称
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}
