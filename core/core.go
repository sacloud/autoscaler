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
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
)

// Core AutoScaler Coreのインスタンス
type Core struct {
	config *Config
}

func newCoreInstance(c *Config) (*Core, error) {
	// TODO バリデーション
	return &Core{config: c}, nil
}

func Start(ctx context.Context, configPath string) error {
	config, err := NewConfigFromPath(configPath)
	if err != nil {
		return err
	}

	instance, err := newCoreInstance(config)
	if err != nil {
		return err
	}

	return instance.run(ctx)
}

func (c *Core) run(ctx context.Context) error {
	errCh := make(chan error)

	// TODO 簡易的な実装、後ほど整理&切り出し
	filename := strings.Replace(defaults.CoreSocketAddr, "unix:", "", -1)
	lis, err := net.Listen("unix", filename)
	if err != nil {
		return fmt.Errorf("starting Core service failed: %s", err)
	}

	server := grpc.NewServer()
	srv := NewScalingService(c)
	request.RegisterScalingServiceServer(server, srv)

	defer func() {
		server.GracefulStop()
		lis.Close() // ignore error
		if _, err := os.Stat(filename); err == nil {
			if err := os.RemoveAll(filename); err != nil {
				log.Printf("cleanup failed: %s\n", err)
			}
		}
	}()

	go func() {
		log.Printf("autoscaler started with: %s\n", lis.Addr().String())
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("core service failed: %s", err)
	case <-ctx.Done():
		log.Println("shutting down with:", ctx.Err())
	}
	return ctx.Err()
}

func (c *Core) Up(ctx *Context) (*Job, error) {
	err := c.handle(ctx)

	// TODO 未実装

	return &Job{
		ID:          c.generateJobID(ctx),
		RequestType: ctx.Request().requestType,
		Status:      request.ScalingJobStatus_JOB_DONE,
	}, err
}

func (c *Core) Down(ctx *Context) (*Job, error) {
	err := c.handle(ctx)

	// TODO 未実装

	return &Job{
		ID:          c.generateJobID(ctx),
		RequestType: ctx.Request().requestType,
		Status:      request.ScalingJobStatus_JOB_DONE,
	}, err
}

func (c *Core) generateJobID(ctx *Context) string {
	return ctx.request.String()
}

func (c *Core) handle(ctx *Context) error {
	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, defaults.HandlerFakeSocketAddr, grpc.WithInsecure())
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
