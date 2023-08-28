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

package builtins

import (
	"log/slog"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/iaas-api-go"
)

// Handler builtinハンドラーをラップし、リクエスト受付時のログ出力を担当するハンドラー
//
// 全ての処理をBuiltinに設定されたハンドラーに委譲する
type Handler struct {
	Builtin handlers.HandlerMeta
}

func (h *Handler) Name() string {
	return h.Builtin.Name()
}

func (h *Handler) Version() string {
	return h.Builtin.Version()
}

func (h *Handler) GetLogger() *slog.Logger {
	return h.Builtin.GetLogger()
}

func (h *Handler) SetLogger(logger *slog.Logger) {
	h.Builtin.SetLogger(logger)
}

func (h *Handler) APICaller() iaas.APICaller {
	if h, ok := h.Builtin.(SakuraCloudAPICaller); ok {
		return h.APICaller()
	}
	return nil
}

func (h *Handler) SetAPICaller(caller iaas.APICaller) {
	if h, ok := h.Builtin.(SakuraCloudAPICaller); ok {
		h.SetAPICaller(caller)
	}
}

func (h *Handler) PreHandle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	logger := h.Builtin.GetLogger()
	logger.Debug(
		"PreHandle() received request",
		slog.String("status", handler.HandleResponse_RECEIVED.String()),
		slog.String("request", req.String()),
	)

	if builtin, ok := h.Builtin.(handlers.PreHandler); ok {
		return builtin.PreHandle(req, sender)
	}

	logger.Debug(
		"PreHandle() ignored request",
		slog.String("status", handler.HandleResponse_IGNORED.String()),
		slog.String("request", req.String()),
	)
	return nil
}

func (h *Handler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	logger := h.Builtin.GetLogger()
	logger.Debug(
		"Handle() received request",
		slog.String("status", handler.HandleResponse_RECEIVED.String()),
		slog.String("request", req.String()),
	)

	if builtin, ok := h.Builtin.(handlers.Handler); ok {
		return builtin.Handle(req, sender)
	}

	logger.Debug(
		"Handle() ignored request",
		slog.String("status", handler.HandleResponse_IGNORED.String()),
		slog.String("request", req.String()),
	)
	return nil
}

func (h *Handler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	logger := h.Builtin.GetLogger()
	logger.Debug(
		"PostHandle() received request",
		slog.String("status", handler.HandleResponse_RECEIVED.String()),
		slog.String("request", req.String()),
	)

	if builtin, ok := h.Builtin.(handlers.PostHandler); ok {
		return builtin.PostHandle(req, sender)
	}

	logger.Debug(
		"PostHandle() ignored request",
		slog.String("status", handler.HandleResponse_IGNORED.String()),
		slog.String("request", req.String()),
	)
	return nil
}
