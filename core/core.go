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

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Core AutoScaler Coreのインスタンス
type Core struct {
	listenAddress string
	config        *Config
	jobs          map[string]*JobStatus
	logger        *log.Logger
}

func newCoreInstance(addr string, c *Config, logger *log.Logger) (*Core, error) {
	return &Core{
		listenAddress: addr,
		config:        c,
		jobs:          make(map[string]*JobStatus),
		logger:        logger,
	}, nil
}

// Start 指定のファイルパスからコンフィグを読み込み、gRPCサーバとしてリッスンを開始する
func Start(ctx context.Context, addr, configPath string, logger *log.Logger) error {
	config, err := NewConfigFromPath(configPath)
	if err != nil {
		return err
	}

	if err := config.Validate(ctx); err != nil {
		return err
	}

	instance, err := newCoreInstance(addr, config, logger)
	if err != nil {
		return err
	}

	return instance.run(ctx)
}

func (c *Core) run(ctx context.Context) error {
	errCh := make(chan error)

	listener, cleanup, err := grpcutil.Listener(&grpcutil.ListenerOption{
		Address: c.listenAddress,
	})
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	srv := NewScalingService(c)
	request.RegisterScalingServiceServer(server, srv)
	reflection.Register(server)

	defer func() {
		server.GracefulStop()
		cleanup()
	}()

	go func() {
		if err := c.logger.Info("message", "autoscaler started", "address", listener.Addr().String()); err != nil {
			errCh <- err
		}
		if err := server.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("core service failed: %s", err)
	case <-ctx.Done():
		c.logger.Info("message", "shutting down", "error", ctx.Err()) // nolint
	}
	return ctx.Err()
}

func (c *Core) Up(ctx *RequestContext) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) Down(ctx *RequestContext) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) currentJob(ctx *RequestContext) *JobStatus {
	job, ok := c.jobs[ctx.JobID()]
	if !ok {
		job = NewJobStatus(ctx.Request(), c.config.AutoScaler.JobCoolDownTime())
		c.jobs[ctx.JobID()] = job
	}
	return job
}

func (c *Core) handle(ctx *RequestContext) (*JobStatus, string, error) {
	job := c.currentJob(ctx)
	if !job.Acceptable() {
		return job, "job is in an unacceptable state", nil
	}

	// 現在のコンテキスト(リクエストスコープ)にjobを保持しておく
	ctx = ctx.WithJobStatus(job)

	//対象リソースグループを取得
	rg, err := c.targetResourceGroup(ctx)
	if err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED)                             // まだ実行前のためCANCELEDを返す
		ctx.Logger().Info("status", request.ScalingJobStatus_JOB_CANCELED, "error", err) // nolint
		return job, "", err
	}

	if err := rg.ValidateActions(ctx.Request().action, c.config.Handlers()); err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED)                             // まだ実行前のためCANCELEDを返す
		ctx.Logger().Info("status", request.ScalingJobStatus_JOB_CANCELED, "error", err) // nolint
		return job, "", err
	}

	go rg.HandleAll(ctx, c.config.APIClient(), c.config.Handlers())

	job.SetStatus(request.ScalingJobStatus_JOB_ACCEPTED)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_ACCEPTED) // nolint
	return job, "", nil
}

func (c *Core) targetResourceGroup(ctx *RequestContext) (*ResourceGroup, error) {
	groupName := ctx.Request().resourceGroupName
	if groupName == "" {
		groupName = defaults.ResourceGroupName
	}

	if groupName == defaults.ResourceGroupName {
		resourceGroups := c.config.Resources.All()
		if len(resourceGroups) > 1 {
			return nil, fmt.Errorf("resource group name %q cannot be specified when multiple groups are defined", defaults.ResourceGroupName)
		}

		return resourceGroups[0], nil
	}

	rg, ok := c.config.Resources.GetOk(groupName)
	if !ok {
		return nil, fmt.Errorf("resource group %q not found", groupName)
	}
	return rg, nil
}
