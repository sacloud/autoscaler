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

package gslb

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ServersHandler GSLB配下のサーバのアタッチ/デタッチを行うためのハンドラ
//
// リクエストされたリソースの直近の親リソースがGSLBの場合に処理を行う
//   - PreHandle: GSLBからのデタッチ
//   - Handle: なにもしない
//   - PostHandle: GSLBへのアタッチ
//
// アタッチ/デタッチは各サーバのEnabledを制御することで行う
// もしGSLBにサーバが1台しか登録されていない場合はサービス停止が発生するため注意が必要
type ServersHandler struct {
	handlers.SakuraCloudFlagCustomizer
	Logger *log.Logger
}

func (h *ServersHandler) Name() string {
	return "gslb-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
}

func (h *ServersHandler) GetLogger() *log.Logger {
	return h.Logger
}

func (h *ServersHandler) PreHandle(req *handler.PreHandleRequest, sender handlers.ResponseSender) error {
	ctx := context.Background()

	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
	}); err != nil {
		return err
	}

	if req.Instruction == handler.ResourceInstructions_UPDATE && h.shouldHandle(req.Desired) {
		// TODO 入力値のバリデーション
		server := req.Desired.GetServer()
		gslb := server.Parent.GetGslb() // バリデーション済みなためnilチェック不要
		if err := h.handle(ctx, req.ScalingJobId, server, gslb, sender, false); err != nil {
			return err
		}
	} else {
		return sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_IGNORED,
		})
	}
	return nil
}

func (h *ServersHandler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	ctx := context.Background()

	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
	}); err != nil {
		return err
	}

	if req.Result == handler.PostHandleRequest_UPDATED && h.shouldHandle(req.Desired) {
		// TODO 入力値のバリデーション
		server := req.Desired.GetServer()
		gslb := server.Parent.GetGslb() // バリデーション済みなためnilチェック不要
		if err := h.handle(ctx, req.ScalingJobId, server, gslb, sender, true); err != nil {
			return err
		}
	} else {
		return sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_IGNORED,
		})
	}
	return nil
}

func (h *ServersHandler) shouldHandle(desired *handler.Resource) bool {
	server := desired.GetServer()
	if server != nil {
		parent := server.Parent
		if parent != nil {
			return parent.GetGslb() != nil
		}
	}
	return false
}

func (h *ServersHandler) handle(ctx context.Context, jobID string, server *handler.Server, gslb *handler.GSLB, sender handlers.ResponseSender, attach bool) error {
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: jobID,
		Status:       handler.HandleResponse_RUNNING,
	}); err != nil {
		return err
	}

	gslbOp := sacloud.NewGSLBOp(h.APIClient())
	current, err := gslbOp.Read(ctx, types.StringID(gslb.Id))
	if err != nil {
		return err
	}

	// バリデーション済み
	shouldUpdate := false
	for _, s := range current.DestinationServers {
		for _, nic := range server.AssignedNetwork {
			if s.IPAddress == nic.IpAddress {
				s.Enabled = types.StringFlag(attach)
				shouldUpdate = true

				if err := sender.Send(&handler.HandleResponse{
					ScalingJobId: jobID,
					Status:       handler.HandleResponse_RUNNING,
					Log:          fmt.Sprintf("found target server: %#v", s),
				}); err != nil {
					return err
				}

				break
			}
		}
	}

	if !shouldUpdate {
		if err := sender.Send(&handler.HandleResponse{
			ScalingJobId: jobID,
			Status:       handler.HandleResponse_DONE,
			Log:          "target server not found",
		}); err != nil {
			return err
		}
		return nil
	}

	if _, err := gslbOp.UpdateSettings(ctx, types.StringID(gslb.Id), &sacloud.GSLBUpdateSettingsRequest{
		HealthCheck:        current.HealthCheck,
		DelayLoop:          current.DelayLoop,
		Weighted:           current.Weighted,
		SorryServer:        current.SorryServer,
		DestinationServers: current.DestinationServers,
		SettingsHash:       current.SettingsHash,
	}); err != nil {
		return err
	}

	return sender.Send(&handler.HandleResponse{
		ScalingJobId: jobID,
		Status:       handler.HandleResponse_DONE,
	})
}
