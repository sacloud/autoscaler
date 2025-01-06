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

package gslb

import (
	"context"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
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
	handlers.HandlerLogger
	*builtins.SakuraCloudAPIClient
}

func NewServersHandler() *ServersHandler {
	return &ServersHandler{
		SakuraCloudAPIClient: &builtins.SakuraCloudAPIClient{},
	}
}

func (h *ServersHandler) Name() string {
	return "gslb-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
}

func (h *ServersHandler) PreHandle(parentCtx context.Context, req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(parentCtx, req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		switch req.Instruction {
		case handler.ResourceInstructions_DELETE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}

			instance := req.Desired.GetServerGroupInstance()
			gslb := instance.Parent.GetGslb()
			return h.deleteServer(ctx, instance, gslb)
		case handler.ResourceInstructions_UPDATE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}

			server := req.Desired.GetServer()
			gslb := server.Parent.GetGslb()
			return h.attachAndDetach(ctx, server, gslb, false)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) PostHandle(parentCtx context.Context, req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(parentCtx, req.ScalingJobId, sender)

	if h.shouldHandle(req.Current) {
		switch req.Result {
		case handler.PostHandleRequest_CREATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}

			instance := req.Current.GetServerGroupInstance()
			gslb := instance.Parent.GetGslb()
			return h.addServer(ctx, instance, gslb)
		case handler.PostHandleRequest_UPDATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			server := req.Current.GetServer()
			gslb := server.Parent.GetGslb() // バリデーション済みなためnilチェック不要
			return h.attachAndDetach(ctx, server, gslb, true)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) shouldHandle(desired *handler.Resource) bool {
	server := desired.GetServer()
	if server != nil {
		parent := server.Parent
		if parent != nil {
			return parent.GetGslb() != nil
		}
	}
	instance := desired.GetServerGroupInstance()
	if instance != nil {
		parent := instance.Parent
		if parent != nil {
			return parent.GetGslb() != nil
		}
	}
	return false
}

func (h *ServersHandler) attachAndDetach(ctx *handlers.HandlerContext, server *handler.Server, gslb *handler.GSLB, attach bool) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	gslbOp := iaas.NewGSLBOp(h.APICaller())
	current, err := gslbOp.Read(ctx, types.StringID(gslb.Id))
	if err != nil {
		return err
	}

	shouldUpdate := false
	targetIPAddress := ""
	for _, s := range current.DestinationServers {
		for _, nic := range server.AssignedNetwork {
			if s.IPAddress == nic.IpAddress {
				s.Enabled = types.StringFlag(attach)
				shouldUpdate = true
				targetIPAddress = s.IPAddress

				if err := ctx.Report(handler.HandleResponse_RUNNING,
					"target server found: {IPAddress: %s}", s.IPAddress); err != nil {
					return err
				}

				break
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

	if _, err := gslbOp.UpdateSettings(ctx, types.StringID(gslb.Id), &iaas.GSLBUpdateSettingsRequest{
		HealthCheck:        current.HealthCheck,
		DelayLoop:          current.DelayLoop,
		Weighted:           current.Weighted,
		SorryServer:        current.SorryServer,
		DestinationServers: current.DestinationServers,
		SettingsHash:       current.SettingsHash,
	}); err != nil {
		return err
	}

	return ctx.Report(handler.HandleResponse_DONE,
		"updated: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress,
	)
}

func (h *ServersHandler) addServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, gslb *handler.GSLB) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	gslbOp := iaas.NewGSLBOp(h.APICaller())
	current, err := gslbOp.Read(ctx, types.StringID(gslb.Id))
	if err != nil {
		return err
	}

	if len(instance.NetworkInterfaces) == 0 {
		return ctx.Report(handler.HandleResponse_IGNORED, "instance has no NICs")
	}
	nic := instance.NetworkInterfaces[0]
	if nic.ExposeInfo == nil {
		return ctx.Report(handler.HandleResponse_IGNORED, "instance.network_interface[0] has no expose info")
	}
	exposeInfo := nic.ExposeInfo

	shouldUpdate := false
	// 存在しなければ追加する
	exist := false
	for _, s := range current.DestinationServers {
		if s.IPAddress == nic.AssignedNetwork.IpAddress {
			exist = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"skipped: Server{IP: %s} already exists on GSLB", s.IPAddress); err != nil {
				return err
			}
			break
		}
	}
	if !exist {
		current.DestinationServers = append(current.DestinationServers, &iaas.GSLBServer{
			IPAddress: nic.AssignedNetwork.IpAddress,
			Enabled:   true,
			Weight:    types.StringNumber(exposeInfo.Weight),
		})
		shouldUpdate = true
		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"added: Server{IP: %s}", nic.AssignedNetwork.IpAddress); err != nil {
			return err
		}
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := gslbOp.UpdateSettings(ctx, current.ID, &iaas.GSLBUpdateSettingsRequest{
			HealthCheck:        current.HealthCheck,
			DelayLoop:          current.DelayLoop,
			Weighted:           current.Weighted,
			SorryServer:        current.SorryServer,
			DestinationServers: current.DestinationServers,
			SettingsHash:       current.SettingsHash,
		})
		if err != nil {
			return err
		}
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updated"); err != nil {
			return err
		}
	}

	return ctx.Report(handler.HandleResponse_DONE)
}

func (h *ServersHandler) deleteServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, gslb *handler.GSLB) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	gslbOp := iaas.NewGSLBOp(h.APICaller())
	current, err := gslbOp.Read(ctx, types.StringID(gslb.Id))
	if err != nil {
		return err
	}

	if len(instance.NetworkInterfaces) == 0 {
		return ctx.Report(handler.HandleResponse_IGNORED, "instance has no NICs")
	}
	nic := instance.NetworkInterfaces[0]

	shouldUpdate := false
	var servers []*iaas.GSLBServer
	for _, s := range current.DestinationServers {
		if s.IPAddress == nic.AssignedNetwork.IpAddress {
			shouldUpdate = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"deleted: Server{IP: %s}", nic.AssignedNetwork.IpAddress); err != nil {
				return err
			}
			continue
		}
		servers = append(servers, s)
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := gslbOp.UpdateSettings(ctx, current.ID, &iaas.GSLBUpdateSettingsRequest{
			HealthCheck:        current.HealthCheck,
			DelayLoop:          current.DelayLoop,
			Weighted:           current.Weighted,
			SorryServer:        current.SorryServer,
			DestinationServers: servers,
			SettingsHash:       current.SettingsHash,
		})
		if err != nil {
			return err
		}
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updated"); err != nil {
			return err
		}
	}

	return ctx.Report(handler.HandleResponse_DONE)
}
