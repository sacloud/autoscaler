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

package stub

import (
	"context"
	"log/slog"
	"os"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/version"
)

// Handler 単体テスト用のスタブハンドラ
type Handler struct {
	PreHandleFunc  func(context.Context, *handler.HandleRequest, handlers.ResponseSender) error
	HandleFunc     func(context.Context, *handler.HandleRequest, handlers.ResponseSender) error
	PostHandleFunc func(context.Context, *handler.PostHandleRequest, handlers.ResponseSender) error
	Logger         *slog.Logger
}

func (h *Handler) Name() string {
	return "stub"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) GetLogger() *slog.Logger {
	if h.Logger != nil {
		return h.Logger
	}
	return log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      false,
		TimeStamp: true,
		Caller:    true,
		Level:     slog.LevelDebug,
	})
}

func (h *Handler) SetLogger(logger *slog.Logger) {
	h.Logger = logger
}

func (h *Handler) PreHandle(parentCtx context.Context, req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if h.PreHandleFunc != nil {
		return h.PreHandleFunc(parentCtx, req, sender)
	}
	return nil
}

func (h *Handler) Handle(parentCtx context.Context, req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if h.HandleFunc != nil {
		return h.HandleFunc(parentCtx, req, sender)
	}
	return nil
}

func (h *Handler) PostHandle(parentCtx context.Context, req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	if h.PostHandleFunc != nil {
		return h.PostHandleFunc(parentCtx, req, sender)
	}
	return nil
}
