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
	"log"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/handlers/server"
	"google.golang.org/grpc"
)

type Handlers []*Handler

var BuiltinHandlers = Handlers{
	// TODO ログの扱いを決めるまでコメントアウトしたまま残しておく
	//{
	//	Type: "logging",
	//	Name: "logging",
	//	BuiltinHandler: &builtins.Handler{
	//		Builtin: &logging.Handler{},
	//	},
	//},
	{
		Type: "server-vertical-scaler",
		Name: "server-vertical-scaler",
		BuiltinHandler: &builtins.Handler{
			Builtin: &server.VerticalScaleHandler{},
		},
	},
	// TODO その他ビルトインを追加
}

// Handler カスタムハンドラーの定義
type Handler struct {
	Type           string          `yaml:"type"`     // ハンドラー種別 TODO: enumにすべきか要検討
	Name           string          `yaml:"name"`     // ハンドラーを識別するための名称 同一Typeで複数のハンドラーが存在する場合が存在するため、Nameで一意に識別する
	Endpoint       string          `yaml:"endpoint"` // カスタムハンドラーの場合にのみ指定
	BuiltinHandler handlers.Server `yaml:"-"`        // ビルトインハンドラーの場合のみ指定
}

func (h *Handler) isBuiltin() bool {
	return h.BuiltinHandler != nil
}

func (h *Handler) Handle(ctx *Context, computed Computed) error {
	// TODO IDが変わるケースに対応するためにhandler呼び出しごとにリフレッシュが必要かも
	if h.isBuiltin() {
		return h.handleBuiltin(ctx, computed)
	}
	return h.handle(ctx, computed)
}

func (h *Handler) handleBuiltin(ctx *Context, computed Computed) error {
	req := ctx.Request()

	if actualHandler, ok := h.BuiltinHandler.(handlers.PreHandler); ok {
		err := actualHandler.PreHandle(&handler.PreHandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Instruction:       computed.Instruction(),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}, &builtinResponseSender{})
		if err != nil {
			return err
		}
	}

	if actualHandler, ok := h.BuiltinHandler.(handlers.Handler); ok {
		err := actualHandler.Handle(&handler.HandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Instruction:       computed.Instruction(),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}, &builtinResponseSender{})
		if err != nil {
			return err
		}
	}

	if actualHandler, ok := h.BuiltinHandler.(handlers.PostHandler); ok {
		err := actualHandler.PostHandle(&handler.PostHandleRequest{
			Source:            req.source,
			Action:            req.action,
			ResourceGroupName: req.resourceGroupName,
			ScalingJobId:      req.ID(),
			Instruction:       computed.Instruction(),
			Current:           computed.Current(),
			Desired:           computed.Desired(),
		}, &builtinResponseSender{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) handle(ctx *Context, computed Computed) error {
	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, h.Endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := handler.NewHandleServiceClient(conn)
	req := ctx.Request()

	preHandleResponse, err := client.PreHandle(ctx, &handler.PreHandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		Instruction:       computed.Instruction(),
		Current:           computed.Current(),
		Desired:           computed.Desired(),
	})
	if err != nil {
		return err
	}
	if err := h.handleHandlerResponse(preHandleResponse); err != nil {
		return err
	}

	handleResponse, err := client.Handle(ctx, &handler.HandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		Instruction:       computed.Instruction(),
		Current:           computed.Current(),
		Desired:           computed.Desired(),
	})
	if err != nil {
		return err
	}
	if err := h.handleHandlerResponse(handleResponse); err != nil {
		return err
	}

	postHandleResponse, err := client.PostHandle(ctx, &handler.PostHandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		Instruction:       computed.Instruction(),
		Current:           computed.Current(),
		Desired:           computed.Desired(),
	})
	if err != nil {
		return err
	}
	if err := h.handleHandlerResponse(postHandleResponse); err != nil {
		return err
	}

	return nil
}

type handlerResponseReceiver interface {
	Recv() (*handler.HandleResponse, error)
}

func (h *Handler) handleHandlerResponse(receiver handlerResponseReceiver) error {
	for {
		stat, err := receiver.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Println("handler replied:", stat.String())
	}
	return nil
}

type builtinResponseSender struct{}

func (s *builtinResponseSender) Send(req *handler.HandleResponse) error {
	log.Println("handler replied:", req.String())
	return nil
}
