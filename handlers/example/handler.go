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

package example

import (
	"context"
	"log/slog"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
)

// Handler ハンドラーの実装例
type Handler struct {
	handlers.HandlerLogger
	listenAddress string
	configPath    string
}

// NewHandler .
func NewHandler(listenAddr string, configPath string, logger *slog.Logger) *Handler {
	return &Handler{
		HandlerLogger: handlers.HandlerLogger{Logger: logger},
		listenAddress: listenAddr,
		configPath:    configPath,
	}
}

// Name ハンドラ名、"autoscaler-handlers-"というプレフィックスをつけない短い名前を返す
func (h *Handler) Name() string {
	return "example"
}

// Version .
func (h *Handler) Version() string {
	return version.FullVersion()
}

// ListenAddress CustomHandlerインターフェースの実装
func (h *Handler) ListenAddress() string {
	return h.listenAddress
}

// ConfigPath CustomHandlerインターフェースの実装
func (h *Handler) ConfigPath() string {
	return h.configPath
}

/*************************************************
 *	必要に応じてPreHandle/Handle/PostHandleを実装する
 *************************************************/

// Handle Coreからのメッセージのハンドリング
func (h *Handler) Handle(parentCtx context.Context, req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(parentCtx, req.ScalingJobId, sender)

	// TODO reqを参照しリクエストを処理すべきか判定

	// 処理すべきリクエストだった場合は受付メッセージ送信
	if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
		return err
	}

	// TODO ここで実際の処理を実装する

	h.GetLogger().Debug("Handle() received request", slog.Any("request", req))

	// 完了メッセージ
	return ctx.Report(handler.HandleResponse_DONE)
}

//// PreHandle Coreからのメッセージのハンドリング
// func (h *Handler) PreHandle(req *handler.PreHandleRequest, sender handlers.ResponseSender) error {
//	// 必要に応じて実装
//	return nil
//}
//

//// PostHandle Coreからのメッセージのハンドリング
// func (h *Handler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
//	// 必要に応じて実装
//	return nil
//}
