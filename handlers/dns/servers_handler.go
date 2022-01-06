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

package dns

import (
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ServersHandler サーバのIPアドレスをAレコード登録/削除するためのハンドラ
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
	return "dns-servers-handler"
}

func (h *ServersHandler) Version() string {
	return version.FullVersion()
}

func (h *ServersHandler) PreHandle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	if h.shouldHandle(req.Desired) {
		switch req.Instruction {
		case handler.ResourceInstructions_DELETE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}

			instance := req.Desired.GetServerGroupInstance()
			dns := instance.Parent.GetDns()
			return h.deleteRecord(ctx, instance, dns)
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
			dns := instance.Parent.GetDns()
			return h.addRecord(ctx, instance, dns)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *ServersHandler) shouldHandle(desired *handler.Resource) bool {
	instance := desired.GetServerGroupInstance()
	if instance != nil {
		parent := instance.Parent
		if parent != nil {
			return parent.GetDns() != nil
		}
	}
	return false
}

func (h *ServersHandler) addRecord(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, dns *handler.DNS) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	dnsOp := sacloud.NewDNSOp(h.APICaller())
	current, err := dnsOp.Read(ctx, types.StringID(dns.Id))
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
	for _, s := range current.Records {
		if s.Type == types.DNSRecordTypes.A && s.Name == exposeInfo.RecordName && s.RData == nic.AssignedNetwork.IpAddress {
			exist = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"skipped: Record{Name:%s, IP:%s} already exists on DNS", s.Name, s.RData); err != nil {
				return err
			}
			break
		}
	}
	if !exist {
		current.Records = append(current.Records, &sacloud.DNSRecord{
			Name:  exposeInfo.RecordName,
			Type:  types.DNSRecordTypes.A,
			RData: nic.AssignedNetwork.IpAddress,
			TTL:   int(nic.ExposeInfo.Ttl),
		})
		shouldUpdate = true
		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"added: Record{Name:%s, IP:%s}", exposeInfo.RecordName, nic.AssignedNetwork.IpAddress); err != nil {
			return err
		}
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := dnsOp.UpdateSettings(ctx, current.ID, &sacloud.DNSUpdateSettingsRequest{
			Records:      current.Records,
			SettingsHash: current.SettingsHash,
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

func (h *ServersHandler) deleteRecord(ctx *handlers.HandlerContext, instance *handler.ServerGroupInstance, dns *handler.DNS) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	dnsOp := sacloud.NewDNSOp(h.APICaller())
	current, err := dnsOp.Read(ctx, types.StringID(dns.Id))
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
	var records []*sacloud.DNSRecord
	for _, r := range current.Records {
		if r.Type == types.DNSRecordTypes.A && r.Name == exposeInfo.RecordName && r.RData == nic.AssignedNetwork.IpAddress {
			shouldUpdate = true
			if err := ctx.Report(handler.HandleResponse_RUNNING,
				"deleted: Record{Name:%s, IP:%s}", exposeInfo.RecordName, nic.AssignedNetwork.IpAddress); err != nil {
				return err
			}
			continue
		}
		records = append(records, r)
	}

	if shouldUpdate {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "updating..."); err != nil {
			return err
		}
		_, err := dnsOp.UpdateSettings(ctx, current.ID, &sacloud.DNSUpdateSettingsRequest{
			Records:      records,
			SettingsHash: current.SettingsHash,
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
