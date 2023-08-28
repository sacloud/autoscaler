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

package localexec

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
)

type Handler struct {
	handlers.HandlerLogger
	listenAddress  string
	configPath     string
	executablePath string
	handlerType    string
}

// NewHandler .
func NewHandler(listenAddr, configPath, executable, handlerType string, logger *slog.Logger) *Handler {
	return &Handler{
		HandlerLogger:  handlers.HandlerLogger{Logger: logger},
		listenAddress:  listenAddr,
		configPath:     configPath,
		executablePath: executable,
		handlerType:    handlerType,
	}
}

// Name ハンドラ名、"autoscaler-handlers-"というプレフィックスをつけない短い名前を返す
func (h *Handler) Name() string {
	return "local-exec"
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

func (h *Handler) PreHandle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if h.handlerType == handlers.HandlerTypePreHandle {
		ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)
		return h.handle(ctx, req.JSON())
	}
	h.GetLogger().Info("PreHandle() received request but ignored")
	return nil
}

func (h *Handler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	if h.handlerType == handlers.HandlerTypeHandle {
		ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)
		return h.handle(ctx, req.JSON())
	}
	h.GetLogger().Info("Handle() received request but ignored")
	return nil
}

func (h *Handler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	if h.handlerType == handlers.HandlerTypePostHandle {
		ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)
		return h.handle(ctx, req.JSON())
	}
	h.GetLogger().Info("PostHandle() received request but ignored")
	return nil
}

func (h *Handler) handle(ctx *handlers.HandlerContext, req []byte) error {
	h.GetLogger().Debug(
		"handle() received request",
		slog.String("status", handler.HandleResponse_RECEIVED.String()),
		slog.String("request", string(req)),
	)

	return h.execute(ctx, req)
}

func (h *Handler) execute(ctx *handlers.HandlerContext, args []byte) error {
	cmd := exec.Command(h.executablePath, string(args)) //nolint: gosec
	argsReader := bytes.NewReader(args)
	cmd.Stdin = argsReader

	output, err := cmd.Output()
	if err != nil {
		wrapped := fmt.Errorf("command %q returned non zero status: %s", h.executablePath, err)
		h.GetLogger().Error(wrapped.Error())
		return wrapped
	}
	h.GetLogger().Info(
		"execute()",
		slog.String("status", handler.HandleResponse_DONE.String()),
		slog.String("output", string(output)),
	)
	return ctx.Report(handler.HandleResponse_DONE, string(output))
}
