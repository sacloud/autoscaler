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
	"log"
	"net"
	"os"
	"strings"

	"github.com/sacloud/autoscaler/defaults"
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
	//対象リソースグループを取得
	resourceGroup, err := c.targetResourceGroup(ctx)
	if err != nil {
		return err
	}

	handlers, err := resourceGroup.Handlers(c.config.Handlers())
	if err != nil {
		return err
	}

	for _, handler := range handlers {
		// desiredはハンドラー処理ごとに再計算する
		allDesired, err := resourceGroup.ComputeAll(ctx, c.config.APIClient())
		if err != nil {
			return err
		}

		if err := handler.Handle(ctx, allDesired); err != nil {
			return err
		}
	}
	return nil
}

func (c *Core) targetResourceGroup(ctx *Context) (*ResourceGroup, error) {
	groupName := ctx.Request().resourceGroupName
	if groupName == defaults.ResourceGroupName {
		// デフォルトではmap内の先頭のリソースグループを返すようにする(yamlでの定義順とは限らない点に注意)
		// TODO 要検討
		for _, v := range c.config.Resources.All() {
			return v, nil
		}
	}

	rg, ok := c.config.Resources.GetOk(groupName)
	if !ok {
		return nil, fmt.Errorf("resource group %q not found", groupName)
	}
	return rg, nil
}
