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
	"io"

	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/handlers/elb"
	"github.com/sacloud/autoscaler/handlers/gslb"
	"github.com/sacloud/autoscaler/handlers/router"
	"github.com/sacloud/autoscaler/handlers/server"
)

type Handlers []*Handler

func BuiltinHandlers() Handlers {
	return Handlers{
		{
			Type: "elb-vertical-scaler",
			Name: "elb-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: &elb.VerticalScaleHandler{},
			},
		},
		{
			Type: "elb-servers-handler",
			Name: "elb-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: &elb.ServersHandler{},
			},
		},
		{
			Type: "gslb-servers-handler",
			Name: "gslb-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: &gslb.ServersHandler{},
			},
		},
		{
			Type: "router-vertical-scaler",
			Name: "router-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: &router.VerticalScaleHandler{},
			},
		},
		{
			Type: "server-vertical-scaler",
			Name: "server-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: &server.VerticalScaleHandler{},
			},
		},
	}
}

// Handler カスタムハンドラーの定義
type Handler struct {
	Type           string               `yaml:"type"`     // ハンドラー種別
	Name           string               `yaml:"name"`     // ハンドラーを識別するための名称 同一Typeで複数のハンドラーが存在する場合が存在するため、Nameで一意に識別する
	Endpoint       string               `yaml:"endpoint"` // カスタムハンドラーの場合にのみ指定
	BuiltinHandler handlers.HandlerMeta `yaml:"-"`        // ビルトインハンドラーの場合のみ指定
	Disabled       bool                 `yaml:"-"`        // ビルトインハンドラーの場合のみ指定
}

func (h *Handler) isBuiltin() bool {
	return h.BuiltinHandler != nil
}

func (h *Handler) PreHandle(ctx *HandlingContext, computed Computed) error {
	if h.isBuiltin() {
		return h.preHandleBuiltin(ctx, computed)
	}
	return h.preHandleExternal(ctx, computed)
}

func (h *Handler) Handle(ctx *HandlingContext, computed Computed) error {
	if h.isBuiltin() {
		return h.handleBuiltin(ctx, computed)
	}
	return h.handleExternal(ctx, computed)
}

func (h *Handler) PostHandle(ctx *HandlingContext, computed Computed) error {
	if h.isBuiltin() {
		return h.postHandleBuiltin(ctx, computed)
	}
	return h.postHandleExternal(ctx, computed)
}

type handleArg struct {
	preHandle  func(request *handler.PreHandleRequest) error
	handle     func(request *handler.HandleRequest) error
	postHandle func(request *handler.PostHandleRequest) error
}

func (h *Handler) handle(ctx *HandlingContext, computed Computed, handleArg *handleArg) error {
	req := ctx.Request()

	if handleArg.preHandle != nil {
		if err := handleArg.preHandle(&handler.PreHandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Instruction:       computed.Instruction(),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}); err != nil {
			return err
		}
	}

	if handleArg.handle != nil {
		if err := handleArg.handle(&handler.HandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Instruction:       computed.Instruction(),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}); err != nil {
			return err
		}
	}

	if handleArg.postHandle != nil {
		if err := handleArg.postHandle(&handler.PostHandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Result:            ctx.ComputeResult(computed),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) preHandleBuiltin(ctx *HandlingContext, computed Computed) error {
	handleArg := &handleArg{}

	if actualHandler, ok := h.BuiltinHandler.(handlers.PreHandler); ok {
		handleArg.preHandle = func(req *handler.PreHandleRequest) error {
			return actualHandler.PreHandle(req, &builtinResponseSender{ctx: ctx})
		}
	}
	return h.handle(ctx, computed, handleArg)
}

func (h *Handler) handleBuiltin(ctx *HandlingContext, computed Computed) error {
	handleArg := &handleArg{}

	if actualHandler, ok := h.BuiltinHandler.(handlers.Handler); ok {
		handleArg.handle = func(req *handler.HandleRequest) error {
			return actualHandler.Handle(req, &builtinResponseSender{ctx: ctx})
		}
	}

	return h.handle(ctx, computed, handleArg)
}

func (h *Handler) postHandleBuiltin(ctx *HandlingContext, computed Computed) error {
	handleArg := &handleArg{}

	if actualHandler, ok := h.BuiltinHandler.(handlers.PostHandler); ok {
		handleArg.postHandle = func(req *handler.PostHandleRequest) error {
			return actualHandler.PostHandle(req, &builtinResponseSender{ctx: ctx})
		}
	}

	return h.handle(ctx, computed, handleArg)
}

func (h *Handler) preHandleExternal(ctx *HandlingContext, computed Computed) error {
	conn, cleanup, err := grpcutil.DialContext(ctx, &grpcutil.DialOption{Destination: h.Endpoint})
	if err != nil {
		return err
	}
	defer cleanup()

	client := handler.NewHandleServiceClient(conn)
	handleArg := &handleArg{
		preHandle: func(req *handler.PreHandleRequest) error {
			res, err := client.PreHandle(ctx, req)
			if err != nil {
				return err
			}
			return h.handleHandlerResponse(ctx, res)
		},
	}
	return h.handle(ctx, computed, handleArg)
}

func (h *Handler) handleExternal(ctx *HandlingContext, computed Computed) error {
	conn, cleanup, err := grpcutil.DialContext(ctx, &grpcutil.DialOption{Destination: h.Endpoint})
	if err != nil {
		return err
	}
	defer cleanup()

	client := handler.NewHandleServiceClient(conn)
	handleArg := &handleArg{
		handle: func(req *handler.HandleRequest) error {
			res, err := client.Handle(ctx, req)
			if err != nil {
				return err
			}
			return h.handleHandlerResponse(ctx, res)
		},
	}
	return h.handle(ctx, computed, handleArg)
}

func (h *Handler) postHandleExternal(ctx *HandlingContext, computed Computed) error {
	conn, cleanup, err := grpcutil.DialContext(ctx, &grpcutil.DialOption{Destination: h.Endpoint})
	if err != nil {
		return err
	}
	defer cleanup()

	client := handler.NewHandleServiceClient(conn)
	handleArg := &handleArg{
		postHandle: func(req *handler.PostHandleRequest) error {
			res, err := client.PostHandle(ctx, req)
			if err != nil {
				return err
			}
			return h.handleHandlerResponse(ctx, res)
		},
	}
	return h.handle(ctx, computed, handleArg)
}

type handlerResponseReceiver interface {
	Recv() (*handler.HandleResponse, error)
}

func (h *Handler) handleHandlerResponse(ctx *HandlingContext, receiver handlerResponseReceiver) error {
	for {
		stat, err := receiver.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := ctx.Logger().Info("status", stat.Status); err != nil {
			return err
		}
		if stat.Log != "" {
			if err := ctx.Logger().Debug("log", stat.Log); err != nil {
				return err
			}
		}
	}
	return nil
}

type builtinResponseSender struct {
	ctx *HandlingContext
}

func (s *builtinResponseSender) Send(res *handler.HandleResponse) error {
	if err := s.ctx.Logger().Info("status", res.Status); err != nil {
		return err
	}
	if res.Log != "" {
		return s.ctx.Logger().Info("log", res.Log)
	}
	return nil
}
