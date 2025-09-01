// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/metrics"
	sacloudotel "github.com/sacloud/autoscaler/otel"
	"github.com/sacloud/autoscaler/request"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Core AutoScaler Coreのインスタンス
type Core struct {
	listenAddress string
	config        *Config
	jobs          map[string]*JobStatus
	logger        *slog.Logger

	mu          sync.RWMutex
	running     bool
	stopping    bool
	shutdownErr error
}

func newCoreInstance(addr string, c *Config, logger *slog.Logger) (*Core, error) {
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
func LoadAndValidate(ctx context.Context, configPath string, strictMode bool, logger *slog.Logger) (*Config, error) {
	if logger == nil {
		logger = log.NewLogger(nil)
	}
	logger.Debug("loading config", slog.String("config-path", configPath))

	config, err := NewConfigFromPath(ctx, configPath, strictMode, logger)
	if err != nil {
		return nil, err
	}

	logger.Debug("validating config")
	if err := config.Validate(ctx); err != nil {
		return nil, err
	}
	return config, nil
}

// New 指定のファイルパスからコンフィグを読み込み、Coreのインスタンスを生成して返すgRPCサーバとしてリッスンを開始する
func New(ctx context.Context, addr, configPath string, strictMode bool, logger *slog.Logger) (*Core, error) {
	ctx, span := sacloudotel.Tracer().Start(ctx, "Core#New",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("sacloud.autoscaler.config.path", configPath)),
		trace.WithAttributes(attribute.Bool("sacloud.autoscaler.config.strict", strictMode)),
	)
	defer span.End()

	logger.Info("starting...")
	return newInstanceFromConfig(ctx, addr, configPath, strictMode, logger)
}

func newInstanceFromConfig(ctx context.Context, addr, configPath string, strictMode bool, logger *slog.Logger) (*Core, error) {
	config, err := LoadAndValidate(ctx, configPath, strictMode, logger)
	if err != nil {
		return nil, err
	}

	instance, err := newCoreInstance(addr, config, logger)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func ResourcesTree(parentCtx context.Context, addr, configPath string, strictMode bool, logger *slog.Logger) (string, error) {
	instance, err := newInstanceFromConfig(parentCtx, addr, configPath, strictMode, logger)
	if err != nil {
		return "", err
	}
	ri := &requestInfo{requestType: requestTypeUnknown}
	ctx := NewRequestContext(parentCtx, ri, logger)
	graph := NewGraph(instance.config.Resources)
	return graph.Tree(ctx, instance.config.APIClient())
}

func (c *Core) Run(ctx context.Context) error {
	return c.run(ctx)
}

func (c *Core) run(ctx context.Context) error {
	errCh := make(chan error)

	// gRPC server
	server, listener, cleanup, err := grpcutil.Server(&grpcutil.ListenerOption{
		Address:    c.listenAddress,
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
		server := metrics.NewServer(exporterConfig.ListenAddress(), c.logger)
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
				c.logger.Error(err.Error())
			}
			exporterListener.Close() //nolint:errcheck
		}()
	}

	go func() {
		c.logger.Info("started", slog.String("address", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("core service failed: %s", err)
	case <-ctx.Done():
		c.logger.Info("shutting down", slog.Any("error", ctx.Err()))
	}
	return c.shutdownErr
}

func (c *Core) Up(ctx *RequestContext) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) Down(ctx *RequestContext) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) Keep(ctx *RequestContext) (*JobStatus, string, error) {
	return c.handle(ctx)
}

func (c *Core) currentJob(ctx *RequestContext) *JobStatus {
	job, ok := c.jobs[ctx.JobID()]
	if !ok {
		job = NewJobStatus(ctx.Request(), c.config.AutoScaler.CoolDown)
		c.jobs[ctx.JobID()] = job
	}
	return job
}

func (c *Core) handle(ctx *RequestContext) (*JobStatus, string, error) {
	traceCtx, span := sacloudotel.Tracer().Start(ctx, "Core#handle",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()
	ctx = ctx.WithContext(traceCtx)

	job := c.currentJob(ctx)
	if c.stopping {
		ctx.Logger().Info(
			"core is shutting down",
			slog.String("status", request.ScalingJobStatus_JOB_IGNORED.String()),
		)
		return job, "core is shutting down", nil
	}

	// 現在のコンテキスト(リクエストスコープ)にjobを保持しておく
	ctx = ctx.WithJobStatus(job)

	// 対象リソースグループを取得
	rds, err := c.targetResourceDef(ctx)
	if err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED) // まだ実行前のためCANCELEDを返す
		ctx.Logger().Info(
			"request has been canceled",
			slog.String("status", request.ScalingJobStatus_JOB_CANCELED.String()),
			slog.Any("error", err))
		return job, "", err
	}

	// さくらのクラウドAPI経由で対象リソース情報を参照し最終更新日時を取得
	lastModifiedAt, err := rds.LastModifiedAt(ctx, c.config.APIClient())
	if err != nil {
		job.SetStatus(request.ScalingJobStatus_JOB_CANCELED) // まだ実行前のためCANCELEDを返す
		ctx.Logger().Info(
			"request has been canceled",
			slog.String("status", request.ScalingJobStatus_JOB_CANCELED.String()),
			slog.Any("error", err),
		)
		return job, "", err
	}

	if !job.Acceptable(ctx.request.requestType, lastModifiedAt) {
		ctx.Logger().Info(
			"job is in an unacceptable state",
			slog.String("status", request.ScalingJobStatus_JOB_IGNORED.String()),
		)
		return job, "job is in an unacceptable state", nil
	}

	job.SetStatus(request.ScalingJobStatus_JOB_ACCEPTED)
	ctx.Logger().Info(
		"request has been accepted",
		slog.String("status", request.ScalingJobStatus_JOB_ACCEPTED.String()),
	)

	c.setRunningStatus(true)

	if ctx.Request().sync {
		rds.HandleAll(ctx, c.config.APIClient(), c.config.Handlers(), func() { c.setRunningStatus(false) })
	} else {
		go rds.HandleAll(ctx, c.config.APIClient(), c.config.Handlers(), func() { c.setRunningStatus(false) })
	}
	return job, "", nil
}

func (c *Core) ResourceName(name string) (string, error) {
	if name == "" || name == defaults.ResourceName {
		if len(c.config.Resources.ResourceNames()) > 1 {
			return "", fmt.Errorf("request parameter 'ResourceName' is required when core's configuration has more than one resource definition")
		}
		name = c.config.Resources[0].Name()
	}
	return name, nil
}

// Stop リクエストの新規受付を停止しつつ現在処理中のUp/Downがあれば終わるまでブロックする
func (c *Core) Stop() error {
	return c.stop(c.config.AutoScaler.ShutdownGracePeriod())
}

func (c *Core) stop(timeout time.Duration) error {
	c.stopping = true

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			err := fmt.Errorf("failed to stop core instance: %s", ctx.Err())
			c.shutdownErr = err
			return err
		case <-ticker.C:
			if !c.isRunning() {
				return nil
			}
		}
	}
}

func (c *Core) setRunningStatus(status bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = status
}

func (c *Core) isRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.running
}

func (c *Core) targetResourceDef(ctx *RequestContext) (ResourceDefinitions, error) {
	name := ctx.Request().resourceName
	defs := c.config.Resources.FilterByResourceName(name)
	if len(defs) > 0 {
		return defs, nil
	}
	return nil, fmt.Errorf("resource %q not found", name)
}
