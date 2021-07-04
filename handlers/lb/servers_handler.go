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

package lb

import (
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ServersHandler ロードバランサ配下のサーバのアタッチ/デタッチを行うためのハンドラ
//
// リクエストされたリソースの直近の親リソースがロードバランサの場合に処理を行う
//   - PreHandle: lbからのデタッチ
//   - Handle: なにもしない
//   - PostHandle: lbへのアタッチ
//
// アタッチ/デタッチは各サーバのEnabledを制御することで行う
// もしLBにサーバが1台しか登録されていない場合はサービス停止が発生するため注意が必要
type ServersHandler struct {
	handlers.HandlerLogger
	*builtins.SakuraCloudAPIClient
}

func NewServersHandler() *ServersHandler {
	return &ServersHandler{
		SakuraCloudAPIClient: &builtins.SakuraCloudAPIClient{},
	}
}

func (h *ServersHandler) Name() string {
	return "load-balancer-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
}

func (h *ServersHandler) PreHandle(req *handler.PreHandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		server := req.Desired.GetServer()
		lb := server.Parent.GetLoadBalancer()
		switch req.Instruction {
		case handler.ResourceInstructions_UPDATE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.handle(ctx, server, lb, false)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		server := req.Desired.GetServer()
		lb := server.Parent.GetLoadBalancer()
		switch req.Result {
		case handler.PostHandleRequest_UPDATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}

			return h.handle(ctx, server, lb, true)
		}
	}
	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) shouldHandle(desired *handler.Resource) bool {
	server := desired.GetServer()
	if server != nil {
		parent := server.Parent
		if parent != nil {
			return parent.GetLoadBalancer() != nil
		}
	}
	return false
}

func (h *ServersHandler) handle(ctx *handlers.HandlerContext, server *handler.Server, lb *handler.LoadBalancer, attach bool) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	lbOp := sacloud.NewLoadBalancerOp(h.APICaller())
	current, err := lbOp.Read(ctx, lb.Zone, types.StringID(lb.Id))
	if err != nil {
		return err
	}

	shouldUpdate := false
	targetIPAddress := ""
	for _, vip := range current.VirtualIPAddresses {
		for _, s := range vip.Servers {
			for _, nic := range server.AssignedNetwork {
				if s.IPAddress == nic.IpAddress {
					s.Enabled = types.StringFlag(attach)
					shouldUpdate = true
					targetIPAddress = s.IPAddress

					if err := ctx.Report(handler.HandleResponse_RUNNING,
						"target server found: %s", s.IPAddress); err != nil {
						return err
					}

					break
				}
			}
		}
	}

	if !shouldUpdate {
		return ctx.Report(handler.HandleResponse_DONE, "target server not found")
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"updating...: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress); err != nil {
		return err
	}

	if _, err := lbOp.UpdateSettings(ctx, lb.Zone, types.StringID(lb.Id), &sacloud.LoadBalancerUpdateSettingsRequest{
		VirtualIPAddresses: current.VirtualIPAddresses,
		SettingsHash:       current.SettingsHash,
	}); err != nil {
		return err
	}

	return ctx.Report(handler.HandleResponse_DONE,
		"updated: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress,
	)
}
