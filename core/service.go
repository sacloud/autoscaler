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

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
)

var _ request.ScalingServiceServer = (*ScalingService)(nil)

type ScalingService struct {
	request.UnimplementedScalingServiceServer
}

func NewScalingService() request.ScalingServiceServer {
	return &ScalingService{}
}

func (s *ScalingService) Up(ctx context.Context, req *request.ScalingRequest) (*request.ScalingResponse, error) {
	log.Println("Core.ScalingService: Up:", req)
	if err := s.handle(ctx, req); err != nil {
		return nil, err
	}
	return &request.ScalingResponse{
		ScalingJobId: "1",
		Status:       request.ScalingJobStatus_DONE,
	}, nil
}

func (s *ScalingService) Down(ctx context.Context, req *request.ScalingRequest) (*request.ScalingResponse, error) {
	log.Println("Core.ScalingService: Down:", req)
	if err := s.handle(ctx, req); err != nil {
		return nil, err
	}
	return &request.ScalingResponse{
		ScalingJobId: "1",
		Status:       request.ScalingJobStatus_DONE,
	}, nil
}

func (s *ScalingService) Status(ctx context.Context, req *request.StatusRequest) (*request.ScalingResponse, error) {
	log.Println("Core.ScalingService: Status:", req)
	return &request.ScalingResponse{
		ScalingJobId: "1",
		Status:       request.ScalingJobStatus_DONE,
	}, nil
}

func (s *ScalingService) handle(ctx context.Context, req *request.ScalingRequest) error {
	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, defaults.HandlerFakeSocketAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := handler.NewHandleServiceClient(conn)
	stream, err := client.Handle(ctx, &handler.HandleRequest{
		Source:            req.Source,
		Action:            req.Action,
		ResourceGroupName: req.ResourceGroupName,
		ScalingJobId:      "1",
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
