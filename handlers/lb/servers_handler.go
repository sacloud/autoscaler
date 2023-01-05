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

package lb

import (
	"fmt"
	"net"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
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

func (h *ServersHandler) PreHandle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		switch req.Instruction {
		case handler.ResourceInstructions_UPDATE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			server := req.Desired.GetServer()
			lb := server.Parent.GetLoadBalancer()
			return h.attachOrDetach(ctx, server, lb, false)
		case handler.ResourceInstructions_DELETE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			instance := req.Desired.GetServerGroupInstance()
			lb := instance.Parent.GetLoadBalancer()
			return h.deleteServer(ctx, instance, lb)
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
			lb := instance.Parent.GetLoadBalancer()
			return h.addServer(ctx, instance, lb)
		case handler.PostHandleRequest_UPDATED:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			server := req.Current.GetServer()
			lb := server.Parent.GetLoadBalancer()
			return h.attachOrDetach(ctx, server, lb, true)
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
	instance := desired.GetServerGroupInstance()
	if instance != nil {
		parent := instance.Parent
		if parent != nil {
			return parent.GetLoadBalancer() != nil
		}
	}
	return false
}

func (h *ServersHandler) attachOrDetach(ctx *handlers.HandlerContext, server *handler.Server, lb *handler.LoadBalancer, attach bool) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	lbOp := iaas.NewLoadBalancerOp(h.APICaller())
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

	if _, err := lbOp.UpdateSettings(ctx, lb.Zone, types.StringID(lb.Id), &iaas.LoadBalancerUpdateSettingsRequest{
		VirtualIPAddresses: current.VirtualIPAddresses,
		SettingsHash:       current.SettingsHash,
	}); err != nil {
		return err
	}
	if err := lbOp.Config(ctx, lb.Zone, types.StringID(lb.Id)); err != nil {
		return err
	}

	return ctx.Report(handler.HandleResponse_DONE,
		"updated: {Enabled:%t, IPAddress:%s}", attach, targetIPAddress,
	)
}

func (h *ServersHandler) sameNetwork(ip1 string, mask1 int, ip2 string) (bool, error) {
	_, net1, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip1, mask1))
	if err != nil {
		return false, err
	}
	return net1.Contains(net.ParseIP(ip2)), nil
}

func (h *ServersHandler) addServerUnderVIP(ctx *handlers.HandlerContext, vip *iaas.LoadBalancerVirtualIPAddress, ip string, port int, healthCheck *handler.ServerGroupInstance_HealthCheck) (bool, error) {
	// すでに同じIPアドレスが登録されていないか??
	exist := false
	for _, s := range vip.Servers {
		if s.IPAddress == ip {
			exist = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"skipped: Server{VIP:%s, IP: %s, Port:%d} already exists on ELB", vip.VirtualIPAddress, ip, port); err != nil {
				return false, err
			}
			break
		}
	}

	if !exist {
		vip.Servers = append(vip.Servers, &iaas.LoadBalancerServer{
			IPAddress: ip,
			Port:      types.StringNumber(port),
			Enabled:   true,
			HealthCheck: &iaas.LoadBalancerServerHealthCheck{
				Protocol:     types.ELoadBalancerHealthCheckProtocol(healthCheck.Protocol),
				Path:         healthCheck.Path,
				ResponseCode: types.StringNumber(healthCheck.StatusCode),
			},
		})
		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"added: Server{VIP:%s, IP: %s, Port:%d}", vip.VirtualIPAddress, ip, port); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (h *ServersHandler) addServersUnderVIPs(ctx *handlers.HandlerContext, current *iaas.LoadBalancer, instance *handler.ServerGroupInstance) (bool, error) {
	shouldUpdate := false
	for _, nic := range instance.NetworkInterfaces {
		if nic.ExposeInfo == nil || len(nic.ExposeInfo.Ports) == 0 || nic.AssignedNetwork == nil {
			continue
		}

		fn := func(ip string, port int) error {
			for _, vip := range h.filteredVIPs(current.VirtualIPAddresses, nic.ExposeInfo) {
				// 同じネットワーク内にありポートが一致するか?
				sameNetwork, err := h.sameNetwork(vip.VirtualIPAddress, current.NetworkMaskLen, ip)
				if err != nil {
					return err
				}
				if !sameNetwork || vip.Port.Int() != port {
					continue
				}

				// vip配下に実サーバを追加
				updated, err := h.addServerUnderVIP(ctx, vip, ip, port, nic.ExposeInfo.HealthCheck)
				if err != nil {
					return err
				}
				// 1件以上が更新されたらshouldUpdateをtrueにする
				if updated {
					shouldUpdate = true
				}
			}
			return nil
		}

		if err := nic.EachIPAndExposedPort(fn); err != nil {
			return false, err
		}
	}
	return shouldUpdate, nil
}

func (h *ServersHandler) addServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, lb *handler.LoadBalancer) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	lbOp := iaas.NewLoadBalancerOp(h.APICaller())
	current, err := lbOp.Read(ctx, lb.Zone, types.StringID(lb.Id))
	if err != nil {
		return err
	}

	if len(instance.NetworkInterfaces) == 0 {
		return ctx.Report(handler.HandleResponse_IGNORED, "instance has no NICs")
	}

	// 実サーバを登録
	shouldUpdate, err := h.addServersUnderVIPs(ctx, current, instance)
	if err != nil {
		return err
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := lbOp.UpdateSettings(ctx, lb.Zone, current.ID, &iaas.LoadBalancerUpdateSettingsRequest{
			VirtualIPAddresses: current.VirtualIPAddresses,
			SettingsHash:       current.SettingsHash,
		})
		if err != nil {
			return err
		}
		if err := lbOp.Config(ctx, lb.Zone, types.StringID(lb.Id)); err != nil {
			return err
		}
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updated"); err != nil {
			return err
		}
	}

	return ctx.Report(handler.HandleResponse_DONE)
}

func (h *ServersHandler) filteredVIPs(vips iaas.LoadBalancerVirtualIPAddresses, exposeInfo *handler.ServerGroupInstance_ExposeInfo) iaas.LoadBalancerVirtualIPAddresses {
	if exposeInfo == nil || len(exposeInfo.Vips) == 0 {
		return vips
	}
	var results iaas.LoadBalancerVirtualIPAddresses
	for _, vip := range vips {
		exist := false
		for _, filter := range exposeInfo.Vips {
			if vip.VirtualIPAddress == filter {
				exist = true
				break
			}
		}
		if exist {
			results = append(results, vip)
		}
	}
	return results
}

func (h *ServersHandler) deleteServer(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, lb *handler.LoadBalancer) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	lbOp := iaas.NewLoadBalancerOp(h.APICaller())
	current, err := lbOp.Read(ctx, lb.Zone, types.StringID(lb.Id))
	if err != nil {
		return err
	}

	if len(instance.NetworkInterfaces) == 0 {
		return ctx.Report(handler.HandleResponse_IGNORED, "instance has no NICs")
	}

	shouldUpdate := false
	for _, nic := range instance.NetworkInterfaces {
		fn := func(ip string, port int) error {
			for _, vip := range h.filteredVIPs(current.VirtualIPAddresses, nic.ExposeInfo) {
				var servers []*iaas.LoadBalancerServer
				for _, server := range vip.Servers {
					if server.IPAddress == ip && server.Port.Int() == port {
						shouldUpdate = true
						if err := ctx.Report(handler.HandleResponse_RUNNING,
							"deleted: Server{VIP:%s, IP: %s, Port:%d}", vip.VirtualIPAddress, ip, port); err != nil {
							return err
						}
						continue
					}
					servers = append(servers, server)
				}
				vip.Servers = servers
			}
			return nil
		}

		if err := nic.EachIPAndExposedPort(fn); err != nil {
			return err
		}
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := lbOp.UpdateSettings(ctx, lb.Zone, current.ID, &iaas.LoadBalancerUpdateSettingsRequest{
			VirtualIPAddresses: current.VirtualIPAddresses,
			SettingsHash:       current.SettingsHash,
		})
		if err != nil {
			return err
		}
		if err := lbOp.Config(ctx, lb.Zone, types.StringID(lb.Id)); err != nil {
			return err
		}
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updated"); err != nil {
			return err
		}
	}

	return ctx.Report(handler.HandleResponse_DONE)
}
