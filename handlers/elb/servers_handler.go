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

package elb

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ServersHandler ELB配下のサーバのアタッチ/デタッチを行うためのハンドラ
//
// リクエストされたリソースの直近の親リソースがELBの場合に処理を行う
//   - PreHandle: ELBからのデタッチ
//   - Handle: なにもしない
//   - PostHandle: ELBへのアタッチ
//
// アタッチ/デタッチは各サーバのEnabledを制御することで行う
// もしELBにサーバが1台しか登録されていない場合はサービス停止が発生するため注意が必要
type ServersHandler struct {
	handlers.SakuraCloudFlagCustomizer
	handlers.HandlerLogger
}

func (h *ServersHandler) Name() string {
	return "elb-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
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
		elb := server.Parent.GetElb() // バリデーション済みなためnilチェック不要
		if err := h.handle(ctx, req.ScalingJobId, server, elb, sender, false); err != nil {
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
		elb := server.Parent.GetElb() // バリデーション済みなためnilチェック不要
		if err := h.handle(ctx, req.ScalingJobId, server, elb, sender, true); err != nil {
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
			elb := parent.GetElb()
			return elb != nil
		}
	}
	return false
}

func (h *ServersHandler) handle(ctx context.Context, jobID string, server *handler.Server, elb *handler.ELB, sender handlers.ResponseSender, attach bool) error {
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: jobID,
		Status:       handler.HandleResponse_RUNNING,
	}); err != nil {
		return err
	}

	elbOp := sacloud.NewProxyLBOp(h.APIClient())
	current, err := elbOp.Read(ctx, types.StringID(elb.Id))
	if err != nil {
		return err
	}

	// バリデーション済み
	shouldUpdate := false
	for _, s := range current.Servers {
		for _, nic := range server.AssignedNetwork {
			if s.IPAddress == nic.IpAddress {
				s.Enabled = attach
				shouldUpdate = true

				if err := sender.Send(&handler.HandleResponse{
					ScalingJobId: jobID,
					Status:       handler.HandleResponse_RUNNING,
					Log:          fmt.Sprintf("found target server: %s", s.IPAddress),
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

	if _, err := elbOp.UpdateSettings(ctx, types.StringID(elb.Id), &sacloud.ProxyLBUpdateSettingsRequest{
		HealthCheck:   current.HealthCheck,
		SorryServer:   current.SorryServer,
		BindPorts:     current.BindPorts,
		Servers:       current.Servers,
		Rules:         current.Rules,
		LetsEncrypt:   current.LetsEncrypt,
		StickySession: current.StickySession,
		Timeout:       current.Timeout,
		Gzip:          current.Gzip,
		SettingsHash:  current.SettingsHash,
	}); err != nil {
		return err
	}

	return sender.Send(&handler.HandleResponse{
		ScalingJobId: jobID,
		Status:       handler.HandleResponse_DONE,
	})
}