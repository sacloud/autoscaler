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

package elb

import (
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
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
	handlers.HandlerLogger
	*builtins.SakuraCloudAPIClient
}

func NewServersHandler() *ServersHandler {
	return &ServersHandler{
		SakuraCloudAPIClient: &builtins.SakuraCloudAPIClient{},
	}
}

func (h *ServersHandler) Name() string {
	return "elb-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
}

func (h *ServersHandler) PreHandle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		switch req.Instruction {
		case handler.ResourceInstructions_UPDATE:
			server := req.Desired.GetServer()
			elb := server.Parent.GetElb()

			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.attachOrDetach(ctx, server, elb, false)
		case handler.ResourceInstructions_DELETE:
			instance := req.Desired.GetServerGroupInstance()
			elb := instance.Parent.GetElb()
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.deleteServer(ctx, instance, elb)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) PostHandle(req *handler.PostHandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Current) {
		switch req.Result {
		case handler.PostHandleRequest_CREATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			instance := req.Current.GetServerGroupInstance()
			elb := instance.Parent.GetElb()
			return h.addServer(ctx, instance, elb)
		case handler.PostHandleRequest_UPDATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			server := req.Current.GetServer()
			elb := server.Parent.GetElb()
			return h.attachOrDetach(ctx, server, elb, true)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
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
	sgInstance := desired.GetServerGroupInstance()
	if sgInstance != nil {
		parent := sgInstance.Parent
		if parent != nil {
			elb := parent.GetElb()
			return elb != nil
		}
	}
	return false
}

func (h *ServersHandler) attachOrDetach(ctx *handlers.HandlerContext, server *handler.Server, elb *handler.ELB, attach bool) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	elbOp := sacloud.NewProxyLBOp(h.APICaller())
	current, err := elbOp.Read(ctx, types.StringID(elb.Id))
	if err != nil {
		return err
	}

	// バリデーション済み
	shouldUpdate := false
	targetIPAddress := ""
	for _, s := range current.Servers {
		for _, nic := range server.AssignedNetwork {
			if s.IPAddress == nic.IpAddress {
				s.Enabled = attach
				shouldUpdate = true
				targetIPAddress = s.IPAddress

				if err := ctx.Report(handler.HandleResponse_RUNNING, "found target server: %s", s.IPAddress); err != nil {
					return err
				}

				break
			}
		}
	}

	if !shouldUpdate {
		return ctx.Report(handler.HandleResponse_IGNORED, "target server not found")
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"updating...: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress); err != nil {
		return err
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

	return ctx.Report(handler.HandleResponse_DONE,
		"updated: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress)
}

func (h *ServersHandler) addServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, elb *handler.ELB) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	elbOp := sacloud.NewProxyLBOp(h.APICaller())
	current, err := elbOp.Read(ctx, types.StringID(elb.Id))
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

	shouldUpdate := false
	fn := func(ip string, port int) error {
		// 存在しなければ追加する
		exist := false
		for _, s := range current.Servers {
			if s.IPAddress == ip && s.Port == port {
				exist = true
				if err := ctx.Report(handler.HandleResponse_RUNNING,
					"skipped: Server{IP: %s, Port:%d} already exists on ELB", ip, port); err != nil {
					return err
				}
				break
			}
		}
		if !exist {
			current.Servers = append(current.Servers, &sacloud.ProxyLBServer{
				IPAddress:   ip,
				Port:        port,
				ServerGroup: nic.ExposeInfo.ServerGroupName,
				Enabled:     true,
			})
			shouldUpdate = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"added: Server{IP: %s, Port:%d}", ip, port); err != nil {
				return err
			}
		}
		return nil
	}
	if err := nic.EachIPAndExposedPort(fn); err != nil {
		return err
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := elbOp.UpdateSettings(ctx, current.ID, &sacloud.ProxyLBUpdateSettingsRequest{
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

func (h *ServersHandler) deleteServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, elb *handler.ELB) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	elbOp := sacloud.NewProxyLBOp(h.APICaller())
	current, err := elbOp.Read(ctx, types.StringID(elb.Id))
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

	shouldUpdate := false
	var servers []*sacloud.ProxyLBServer
	fn := func(ip string, port int) error {
		for _, s := range current.Servers {
			if s.IPAddress == ip && s.Port == port {
				shouldUpdate = true
				if err := ctx.Report(handler.HandleResponse_RUNNING,
					"deleted: Server{IP: %s, Port:%d}", ip, port); err != nil {
					return err
				}
				continue
			}
			servers = append(servers, s)
		}
		return nil
	}
	if err := nic.EachIPAndExposedPort(fn); err != nil {
		return err
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := elbOp.UpdateSettings(ctx, current.ID, &sacloud.ProxyLBUpdateSettingsRequest{
			HealthCheck:   current.HealthCheck,
			SorryServer:   current.SorryServer,
			BindPorts:     current.BindPorts,
			Servers:       servers,
			Rules:         current.Rules,
			LetsEncrypt:   current.LetsEncrypt,
			StickySession: current.StickySession,
			Timeout:       current.Timeout,
			Gzip:          current.Gzip,
			SettingsHash:  current.SettingsHash,
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
