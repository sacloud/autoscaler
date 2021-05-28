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

	"github.com/sacloud/autoscaler/handlers/server"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/logging"
	"google.golang.org/grpc"
)

type Handlers []*Handler

var BuiltinHandlers = Handlers{
	{
		Type:           "logging",
		Name:           "logging",
		BuiltinHandler: &logging.Handler{},
	},
	{
		Type:           "server-vertical-scaler",
		Name:           "server-vertical-scaler",
		BuiltinHandler: &server.VerticalScaleHandler{},
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

func (h *Handler) Handle(ctx *Context, allDesired []Desired) error {
	var resources []*handler.Resource
	for _, desired := range allDesired {
		resource := desired.ToRequest()
		if resource != nil {
			resources = append(resources, resource)
		}
	}

	if h.isBuiltin() {
		return h.handleBuiltin(ctx, resources)
	}
	return h.handle(ctx, resources)
}

func (h *Handler) handleBuiltin(ctx *Context, resources []*handler.Resource) error {
	req := ctx.Request()
	return h.BuiltinHandler.Handle(&handler.HandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		Resources:         resources,
	}, &builtinResponseSender{})
}

func (h *Handler) handle(ctx *Context, resources []*handler.Resource) error {
	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, h.Endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := handler.NewHandleServiceClient(conn)
	req := ctx.Request()
	stream, err := client.Handle(ctx, &handler.HandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		// サーバが存在するパターン
		Resources: resources,
	})
	if err != nil {
		return err
	}
	for {
		stat, err := stream.Recv()
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
