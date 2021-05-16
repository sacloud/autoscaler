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

package fake

import (
	"log"
	"time"

	"github.com/sacloud/autoscaler/handler"
)

var _ handler.HandleServiceServer = (*HandleService)(nil)

type HandleService struct {
	handler.UnimplementedHandleServiceServer
}

func NewFakeHandlerService() handler.HandleServiceServer {
	return &HandleService{}
}

func (h *HandleService) Handle(req *handler.HandleRequest, server handler.HandleService_HandleServer) error {
	log.Printf("received: %s", req.String())

	// 受付メッセージ送信
	if err := server.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
		Log:          "",
	}); err != nil {
		return err
	}

	// 数回ほど処理中ステータスを返しておく
	for i := 0; i < 3; i++ {
		if err := server.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_RUNNING,
			Log:          "",
		}); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}

	return server.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
		Log:          "",
	})
}
