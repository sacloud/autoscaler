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
	"io"
	"log"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/fake"
	"google.golang.org/grpc"
)

type Handlers []*Handler

var BuiltinHandlers = Handlers{
	{
		Type:           "fake",
		Name:           "fake",
		Endpoint:       "unix:autoscaler-handlers-fake.sock", // ビルトインの場合は後ほどstartBuiltinHandlersを実行した際に設定される
		BuiltinHandler: &fake.Handler{},
	},
	// TODO その他ビルトインを追加
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

// Handler カスタムハンドラーの定義
type Handler struct {
	Type           string `yaml:"type"` // ハンドラー種別 TODO: enumにすべきか要検討
	Name           string `yaml:"name"` // ハンドラーを識別するための名称
	Endpoint       string `yaml:"endpoint"`
	BuiltinHandler handlers.Server
}

func (h *Handler) isBuiltin() bool {
	return h.BuiltinHandler != nil
}

func (h *Handler) Handle(ctx *Context) error {
	if h.isBuiltin() {
		return h.handleBuiltin(ctx)
	}
	return h.handle(ctx)
}

func (h *Handler) handleBuiltin(ctx *Context) error {
	req := ctx.Request()
	return h.BuiltinHandler.Handle(&handler.HandleRequest{
		Source:            req.source,
		Action:            req.action,
		ResourceGroupName: req.resourceGroupName,
		ScalingJobId:      req.ID(),
		// サーバが存在するパターン
		Resources: []*handler.Resource{
			{
				Resource: &handler.Resource_Server{
					Server: &handler.Server{
						Status: handler.ResourceStatus_RUNNING,
						Id:     "123456789012",
						AssignedNetwork: &handler.NetworkInfo{
							IpAddress: "192.0.2.11",
							Netmask:   24,
							Gateway:   "192.0.2.1",
						},
						Core:          2,
						Memory:        4,
						DedicatedCpu:  false,
						PrivateHostId: "",
					}},
			},
		},
	}, &builtinResponseSender{})
}

func (h *Handler) handle(ctx *Context) error {
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
		Resources: []*handler.Resource{
			{
				Resource: &handler.Resource_Server{
					Server: &handler.Server{
						Status: handler.ResourceStatus_RUNNING,
						Id:     "123456789012",
						AssignedNetwork: &handler.NetworkInfo{
							IpAddress: "192.0.2.11",
							Netmask:   24,
							Gateway:   "192.0.2.1",
						},
						Core:          2,
						Memory:        4,
						DedicatedCpu:  false,
						PrivateHostId: "",
					}},
			},
		},
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
