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
	"net"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/metrics"
	"github.com/sacloud/autoscaler/request"
	health "google.golang.org/grpc/health/grpc_health_v1"
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
	metrics.InitErrorCount("core")
	metrics.InitErrorCount("core_to_handlers")

	return &Core{
		listenAddress: addr,
		config:        c,
		jobs:          make(map[string]*JobStatus),
		logger:        logger,
	}, nil
}

// LoadAndValidate 指定のファイルパスからコンフィグを読み込み、バリデーションを行う
func LoadAndValidate(ctx context.Context, configPath string, logger *log.Logger) (*Config, error) {
	config, err := NewConfigFromPath(configPath)
	if err != nil {
		return nil, err
	}

	if err := config.Validate(ctx); err != nil {
		return nil, err
	}
	return config, nil
}

// Start 指定のファイルパスからコンフィグを読み込み、gRPCサーバとしてリッスンを開始する
func Start(ctx context.Context, addr, configPath string, logger *log.Logger) error {
	instance, err := newInstanceFromConfig(ctx, addr, configPath, logger)
	if err != nil {
		return err
	}
	return instance.run(ctx)
}

func newInstanceFromConfig(ctx context.Context, addr, configPath string, logger *log.Logger) (*Core, error) {
	config, err := LoadAndValidate(ctx, configPath, logger)
	if err != nil {
		return nil, err
	}

	instance, err := newCoreInstance(addr, config, logger)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func ResourcesTree(parentCtx context.Context, addr, configPath string, logger *log.Logger) (string, error) {
	instance, err := newInstanceFromConfig(parentCtx, addr, configPath, logger)
	if err != nil {
		return "", err
	}
	ri := &requestInfo{requestType: requestTypeUnknown}
	ctx := NewRequestContext(parentCtx, ri, instance.config.AutoScaler.HandlerTLSConfig, logger)
	graph := NewGraph(instance.config.Resources)
	return graph.Tree(ctx, instance.config.APIClient())
}

func (c *Core) run(ctx context.Context) error {
	errCh := make(chan error)

	// gRPC server
	server, listener, cleanup, err := grpcutil.Server(&grpcutil.ListenerOption{
		Address:    c.listenAddress,
		TLSConfig:  c.config.AutoScaler.ServerTLSConfig,
		ServerOpts: grpcutil.ServerErrorCountInterceptor("core"),
	})
	if err != nil {
		return err
	}
	srv := NewScalingService(c)
	request.RegisterScalingServiceServer(server, srv)
	health.RegisterHealthServer(server, srv)
	reflection.Register(server)

	defer func() {
		server.GracefulStop()
		cleanup()
	}()

	// metrics server
	if c.config.AutoScaler.ExporterEnabled() {
		exporterConfig := c.config.AutoScaler.ExporterConfig
		server := metrics.NewServer(exporterConfig.ListenAddress(), exporterConfig.TLSConfig, c.logger)
		exporterListener, err := net.Listen("tcp", exporterConfig.ListenAddress())
		if err != nil {
			return err
		}

		go func() {
			if err := server.Serve(exporterListener); err != nil {
				errCh <- err
			}
		}()
		defer func() {
			if err := server.Shutdown(ctx); err != nil {
				c.logger.Error("error", err) // nolint
			}
			exporterListener.Close() // nolint
		}()
	}

	go func() {
		if err := c.logger.Info("message", "autoscaler core started", "address", listener.Addr().String()); err != nil {
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
		ctx.Logger().Info("status", request.ScalingJobStatus_JOB_IGNORED, "message", "job is in an unacceptable state") // nolint
		return job, "job is in an unacceptable state", nil
	}

	// 現在のコンテキスト(リクエストスコープ)にjobを保持しておく
	ctx = ctx.WithJobStatus(job)

	//対象リソースグループを取得
	rds, err := c.targetResourceDef(ctx)
	if err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED)                             // まだ実行前のためCANCELEDを返す
		ctx.Logger().Info("status", request.ScalingJobStatus_JOB_CANCELED, "error", err) // nolint
		return job, "", err
	}

	go rds.HandleAll(ctx, c.config.APIClient(), c.config.Handlers())

	job.SetStatus(request.ScalingJobStatus_JOB_ACCEPTED)
	ctx.Logger().Info("status", request.ScalingJobStatus_JOB_ACCEPTED) // nolint
	return job, "", nil
}

func (c *Core) targetResourceDef(ctx *RequestContext) (ResourceDefinitions, error) {
	name := ctx.Request().resourceName
	if name == "" {
		name = defaults.ResourceName
	}

	if name == defaults.ResourceName {
		return ResourceDefinitions{c.config.Resources[0]}, nil
	}

	// TODO ResourceNameからResourceDefを探す処理を実装
	return ResourceDefinitions{c.config.Resources[0]}, nil
}
