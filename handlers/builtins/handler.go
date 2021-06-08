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

package builtins

import (
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/log"
)

// Handler builtinハンドラーをラップし、リクエスト受付時のログ出力を担当するハンドラー
//
// 全ての処理をBuiltinに設定されたハンドラーに委譲する
type Handler struct {
	Builtin handlers.Server
}

func (h *Handler) Name() string {
	return h.Builtin.Name()
}

func (h *Handler) Version() string {
	return h.Builtin.Version()
}

func (h *Handler) GetLogger() *log.Logger {
	return h.Builtin.GetLogger()
}

func (h *Handler) PreHandle(req *handler.PreHandleRequest, sender handlers.ResponseSender) error {
	if builtin, ok := h.Builtin.(handlers.PreHandler); ok {
		if err := h.GetLogger().Info("message", "PreHandle request received", "request", req.String()); err != nil {
			return err
		}
		return builtin.PreHandle(req, sender)
	}

	return h.GetLogger().Info("message", "PreHandle request ignored", "request", req.String())
}

func (h *Handler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if builtin, ok := h.Builtin.(handlers.Handler); ok {
		if err := h.GetLogger().Info("message", "Handle request received", "request", req.String()); err != nil {
			return err
		}
		return builtin.Handle(req, sender)
	}

	return h.GetLogger().Info("message", "Handle request ignored", "request", req.String())
}

func (h *Handler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	if builtin, ok := h.Builtin.(handlers.PostHandler); ok {
		if err := h.GetLogger().Info("message", "PostHandle request received", "request", req.String()); err != nil {
			return err
		}
		return builtin.PostHandle(req, sender)
	}

	return h.GetLogger().Info("message", "PostHandle request ignored", "request", req.String())
}
