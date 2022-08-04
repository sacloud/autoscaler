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

package handlers

import (
	"context"

	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/metrics"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var _ handler.HandleServiceServer = (*handleService)(nil)

// handleService ハンドラ向けgRPCサーバの実装
type handleService struct {
	handler.UnimplementedHandleServiceServer
	Handler CustomHandler
	conf    *Config
}

func (h *handleService) listenAndServe(ctx context.Context) error {
	errCh := make(chan error)

	metrics.InitErrorCount("core")
	opts := &grpcutil.ListenerOption{
		Address:    h.Handler.ListenAddress(),
		ServerOpts: grpcutil.ServerErrorCountInterceptor("handlers"),
	}

	grpcServer, listener, cleanup, err := grpcutil.Server(opts)
	if err != nil {
		h.Handler.GetLogger().Fatal("fatal", err)
		return err // 到達しない
	}

	handler.RegisterHandleServiceServer(grpcServer, h)
	health.RegisterHealthServer(grpcServer, h)
	reflection.Register(grpcServer)

	defer func() {
		grpcServer.GracefulStop()
		cleanup()
	}()

	go func() {
		if err := h.Handler.GetLogger().Info("message", "started", "address", listener.Addr().String()); err != nil {
			errCh <- err
		}
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		h.Handler.GetLogger().Info("message", "shutting down", "error", ctx.Err()) // nolint
	}
	return ctx.Err()
}

func (h *handleService) PreHandle(req *handler.HandleRequest, server handler.HandleService_PreHandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "PreHandle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(PreHandler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PreHandle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}

func (h *handleService) Handle(req *handler.HandleRequest, server handler.HandleService_HandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "Handle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(Handler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.Handle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}

func (h *handleService) PostHandle(req *handler.PostHandleRequest, server handler.HandleService_PostHandleServer) error {
	logger := h.Handler.GetLogger().With(
		"scaling-job-id", req.ScalingJobId,
		"step", "PostHandle",
		"handler", handlerFullName(h.Handler),
	)

	if impl, ok := h.Handler.(PostHandler); ok {
		if err := logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PostHandle(req, server)
	}

	if err := logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return logger.Debug("request", req.String())
}

// Check gRPCヘルスチェックの実装
func (h *handleService) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

// Watch gRPCヘルスチェックの実装
func (h *handleService) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}
