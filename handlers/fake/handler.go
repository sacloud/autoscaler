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
	"time"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
)

type Handler struct {
	handlers.HandlerLogger
}

func (h *Handler) Name() string {
	return "fake"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	// 受付メッセージ送信
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
	}); err != nil {
		return err
	}

	// 数回ほど処理中ステータスを返しておく
	for i := 0; i < 3; i++ {
		if err := sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_RUNNING,
		}); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}

	// 完了メッセージ
	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
	})
}
