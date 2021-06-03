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
	jobs   map[string]*JobStatus
}

func newCoreInstance(c *Config) (*Core, error) {
	// TODO バリデーションの実装
	return &Core{
		config: c,
		jobs:   make(map[string]*JobStatus),
	}, nil
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

func (c *Core) Up(ctx *Context) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) Down(ctx *Context) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) currentJob(ctx *Context) *JobStatus {
	job, ok := c.jobs[ctx.JobID()]
	if !ok {
		job = NewJobStatus(ctx.Request(), c.config.AutoScaler.JobCoolingTime())
		c.jobs[ctx.JobID()] = job
	}
	return job
}

func (c *Core) handle(ctx *Context) (*JobStatus, string, error) {
	job := c.currentJob(ctx)
	if !job.Acceptable() {
		return job, "job is in an unacceptable state", nil
	}

	// 現在のコンテキスト(リクエストスコープ)にjobを保持しておく
	ctx = ctx.WithJobStatus(job)

	//対象リソースグループを取得
	rg, err := c.targetResourceGroup(ctx)
	if err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED) // まだ実行前のためCANCELEDを返す
		return job, "", err
	}

	if err := rg.ValidateHandlerFilters(c.config.Handlers()); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED) // まだ実行前のためCANCELEDを返す
		return job, "", err
	}

	go rg.HandleAll(ctx, c.config.APIClient(), c.config.Handlers())

	job.SetStatus(request.ScalingJobStatus_JOB_ACCEPTED)
	return job, "", nil
}

func (c *Core) targetResourceGroup(ctx *Context) (*ResourceGroup, error) {
	groupName := ctx.Request().resourceGroupName
	if groupName == "" {
		groupName = defaults.ResourceGroupName
	}

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
