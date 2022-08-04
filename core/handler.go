// Copyright 2021-2022 The sacloud/autoscaler Authors
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
	"github.com/sacloud/autoscaler/handlers/dns"
	"github.com/sacloud/autoscaler/handlers/elb"
	"github.com/sacloud/autoscaler/handlers/gslb"
	"github.com/sacloud/autoscaler/handlers/lb"
	"github.com/sacloud/autoscaler/handlers/router"
	"github.com/sacloud/autoscaler/handlers/server"
)

// Handlers Handlerのリスト
type Handlers []*Handler

// BuiltinHandlers ビルトインハンドラのリスト
//
// この段階では各ハンドラにAPIクライアントは注入されない
// Config.Handlersも参照
func BuiltinHandlers() Handlers {
	return Handlers{
		{
			Name: "dns-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: dns.NewServersHandler(),
			},
		},
		{
			Name: "elb-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: elb.NewVerticalScaleHandler(),
			},
		},
		{
			Name: "elb-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: elb.NewServersHandler(),
			},
		},
		{
			Name: "gslb-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: gslb.NewServersHandler(),
			},
		},
		{
			Name: "load-balancer-servers-handler",
			BuiltinHandler: &builtins.Handler{
				Builtin: lb.NewServersHandler(),
			},
		},
		{
			Name: "router-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: router.NewVerticalScaleHandler(),
			},
		},
		{
			Name: "server-horizontal-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: server.NewHorizontalScaleHandler(),
			},
		},
		{
			Name: "server-vertical-scaler",
			BuiltinHandler: &builtins.Handler{
				Builtin: server.NewVerticalScaleHandler(),
			},
		},
	}
}

// Handler カスタムハンドラーの定義
type Handler struct {
	Name           string               `yaml:"name"`     // ハンドラーを識別するための名称
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
	preHandle  func(request *handler.HandleRequest) error
	handle     func(request *handler.HandleRequest) error
	postHandle func(request *handler.PostHandleRequest) error
}

func (h *Handler) handle(ctx *HandlingContext, computed Computed, handleArg *handleArg) error {
	req := ctx.Request()

	if handleArg.preHandle != nil {
		if err := handleArg.preHandle(&handler.HandleRequest{
			Source:           req.source,
			ResourceName:     req.resourceName,
			ScalingJobId:     req.ID(),
			Instruction:      computed.Instruction(),
			SetupGracePeriod: uint32(computed.SetupGracePeriod()),
			Desired:          computed.Desired(),
		}); err != nil {
			return err
		}
	}

	if handleArg.handle != nil {
		if err := handleArg.handle(&handler.HandleRequest{
			Source:           req.source,
			ResourceName:     req.resourceName,
			ScalingJobId:     req.ID(),
			Instruction:      computed.Instruction(),
			SetupGracePeriod: uint32(computed.SetupGracePeriod()),
			Desired:          computed.Desired(),
		}); err != nil {
			return err
		}
	}

	if handleArg.postHandle != nil {
		if err := handleArg.postHandle(&handler.PostHandleRequest{
			Source:           req.source,
			ResourceName:     req.resourceName,
			ScalingJobId:     req.ID(),
			Result:           ctx.ComputeResult(computed),
			Current:          computed.Current(),
			SetupGracePeriod: uint32(computed.SetupGracePeriod()),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) preHandleBuiltin(ctx *HandlingContext, computed Computed) error {
	handleArg := &handleArg{}

	if actualHandler, ok := h.BuiltinHandler.(handlers.PreHandler); ok {
		handleArg.preHandle = func(req *handler.HandleRequest) error {
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
	opt := &grpcutil.DialOption{
		Destination: h.Endpoint,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("core_to_handlers"),
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, opt)
	if err != nil {
		return err
	}
	defer cleanup()

	client := handler.NewHandleServiceClient(conn)
	handleArg := &handleArg{
		preHandle: func(req *handler.HandleRequest) error {
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
	opt := &grpcutil.DialOption{
		Destination: h.Endpoint,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("core_to_handlers"),
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, opt)
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
	opt := &grpcutil.DialOption{
		Destination: h.Endpoint,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("core_to_handlers"),
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, opt)
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
		kvs := []interface{}{"status", stat.Status}
		if stat.Log != "" {
			kvs = append(kvs, "log", stat.Log)
		}

		handleHandlerResponseStatus(ctx, stat.Status)

		if stat.Status == handler.HandleResponse_IGNORED && stat.Log == "" {
			if err := ctx.Logger().Debug(kvs...); err != nil {
				return err
			}
		}
		if err := ctx.Logger().Info(kvs...); err != nil {
			return err
		}
	}
	return nil
}

func handleHandlerResponseStatus(ctx *HandlingContext, status handler.HandleResponse_Status) {
	// いずれかのハンドラが一度でもRUNNING/DONEを返したらハンドラ処理済みとみなす
	if status == handler.HandleResponse_RUNNING || status == handler.HandleResponse_DONE {
		ctx.RequestContext.handled = true
	}
}

type builtinResponseSender struct {
	ctx *HandlingContext
}

func (s *builtinResponseSender) Send(res *handler.HandleResponse) error {
	kvs := []interface{}{"status", res.Status}
	if res.Log != "" {
		kvs = append(kvs, "log", res.Log)
	}

	handleHandlerResponseStatus(s.ctx, res.Status)

	if res.Status == handler.HandleResponse_IGNORED && res.Log == "" {
		return s.ctx.Logger().Debug(kvs...)
	}
	return s.ctx.Logger().Info(kvs...)
}
