// Copyright 2021-2023 The sacloud/autoscaler Authors
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
	"log/slog"

	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var _ request.ScalingServiceServer = (*ScalingService)(nil)
var _ health.HealthServer = (*ScalingService)(nil)

type ScalingService struct {
	request.UnimplementedScalingServiceServer
	instance *Core
}

func NewScalingService(instance *Core) *ScalingService {
	return &ScalingService{instance: instance}
}

func (s *ScalingService) Up(ctx context.Context, req *request.ScalingRequest) (*request.ScalingResponse, error) {
	logger := s.instance.logger.With(
		"request", requestTypeUp.String(),
		"resource", req.ResourceName,
	)
	if req.DesiredStateName != "" {
		logger = logger.With("desired", req.DesiredStateName)
	}
	logger.Info("request received")
	logger.Debug("", slog.Any("request", req))

	resourceName, err := s.instance.ResourceName(req.ResourceName)
	if err != nil {
		return nil, err
	}

	// リクエストには即時応答を返しつつバックグラウンドでジョブを実行するために引数のctxは引き継がない
	serviceCtx := NewRequestContext(context.Background(), &requestInfo{
		requestType:      requestTypeUp,
		source:           req.Source,
		resourceName:     resourceName,
		desiredStateName: req.DesiredStateName,
		sync:             req.Sync,
	}, s.instance.logger)
	job, message, err := s.instance.Up(serviceCtx)
	if err != nil {
		return nil, err
	}
	return &request.ScalingResponse{
		ScalingJobId: job.ID(),
		Status:       job.Status(),
		Message:      message,
	}, nil
}

func (s *ScalingService) Down(ctx context.Context, req *request.ScalingRequest) (*request.ScalingResponse, error) {
	logger := s.instance.logger.With(
		"request", requestTypeUp.String(),
		"resource", req.ResourceName,
	)
	if req.DesiredStateName != "" {
		logger = logger.With("desired", req.DesiredStateName)
	}
	logger.Info("request received")
	logger.Debug("", slog.Any("request", req))

	resourceName, err := s.instance.ResourceName(req.ResourceName)
	if err != nil {
		return nil, err
	}

	// リクエストには即時応答を返しつつバックグラウンドでジョブを実行するために引数のctxは引き継がない
	serviceCtx := NewRequestContext(context.Background(), &requestInfo{
		requestType:      requestTypeDown,
		source:           req.Source,
		resourceName:     resourceName,
		desiredStateName: req.DesiredStateName,
		sync:             req.Sync,
	}, s.instance.logger)
	job, message, err := s.instance.Down(serviceCtx)
	if err != nil {
		return nil, err
	}
	return &request.ScalingResponse{
		ScalingJobId: job.ID(),
		Status:       job.Status(),
		Message:      message,
	}, nil
}

// Check gRPCヘルスチェックの実装
func (s *ScalingService) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

// Watch gRPCヘルスチェックの実装
func (s *ScalingService) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}
