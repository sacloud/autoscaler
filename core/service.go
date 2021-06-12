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

	"github.com/sacloud/autoscaler/request"
)

var _ request.ScalingServiceServer = (*ScalingService)(nil)

type ScalingService struct {
	request.UnimplementedScalingServiceServer
	instance *Core
}

func NewScalingService(instance *Core) request.ScalingServiceServer {
	return &ScalingService{instance: instance}
}

func (s *ScalingService) Up(ctx context.Context, req *request.ScalingRequest) (*request.ScalingResponse, error) {
	if err := s.instance.logger.Info("message", "request received", "request-type", requestTypeUp); err != nil {
		return nil, err
	}
	if err := s.instance.logger.Debug("request", req); err != nil {
		return nil, err
	}

	// リクエストには即時応答を返しつつバックグラウンドでジョブを実行するために引数のctxは引き継がない
	serviceCtx := NewRequestContext(context.Background(), &requestInfo{
		requestType:       requestTypeUp,
		source:            req.Source,
		action:            req.Action,
		resourceGroupName: req.ResourceGroupName,
		desiredStateName:  req.DesiredStateName,
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
	if err := s.instance.logger.Info("message", "request received", "request-type", requestTypeDown); err != nil {
		return nil, err
	}
	if err := s.instance.logger.Debug("request", req); err != nil {
		return nil, err
	}

	// リクエストには即時応答を返しつつバックグラウンドでジョブを実行するために引数のctxは引き継がない
	serviceCtx := NewRequestContext(context.Background(), &requestInfo{
		requestType:       requestTypeDown,
		source:            req.Source,
		action:            req.Action,
		resourceGroupName: req.ResourceGroupName,
		desiredStateName:  req.DesiredStateName,
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
