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
	"log"
)

type Handlers []*Handler

var BuiltinHandlers = Handlers{
	{
		Type:     "server_vertical_scaler",
		Name:     "server_vertical_scaler",
		Endpoint: "server_vertical_scaler.sock", // ビルトインの場合は後ほどstartBuiltinHandlersを実行した際に設定される
	},
	// TODO その他ビルトインを追加
}

// Handler カスタムハンドラーの定義
type Handler struct {
	Type     string `yaml:"type"` // ハンドラー種別 TODO: enumにすべきか要検討
	Name     string `yaml:"name"` // ハンドラーを識別するための名称
	Endpoint string `yaml:"endpoint"`
}

func (h *Handler) isBuiltin() bool {
	return h.Type == "server_vertical_scaler" // TODO ビルトインを増やす際に修正
}

func startBuiltinHandlers(ctx context.Context, handlers Handlers) error {
	// TODO ソケットのパスを受け取れるように修正
	for _, h := range handlers {
		if h.isBuiltin() {
			// TODO ビルトインの開始
			log.Println("startBuiltinHandlers is not implemented")
		}
	}
	return nil
}
