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

package core

import (
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

type ResourceServerGroupInstance struct {
	*ResourceBase

	apiClient    iaas.APICaller
	server       *iaas.Server
	zone         string
	def          *ResourceDefServerGroup
	instruction  handler.ResourceInstructions
	indexInGroup int // グループ内でのインデックス、値の算出に用いる

	parent Resource
}

func (r *ResourceServerGroupInstance) String() string {
	if r == nil || r.server == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, Zone: %s, ID: %s, Name: %s}", r.Type(), r.zone, r.server.ID, r.server.Name)
}

func (r *ResourceServerGroupInstance) Parent() Resource {
	return r.parent
}

func (r *ResourceServerGroupInstance) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
		r.instruction = handler.ResourceInstructions_NOOP
	}

	var parentComputed Computed
	if r.parent != nil {
		c, err := r.parent.Compute(ctx, refresh)
		if err != nil {
			return nil, err
		}
		parentComputed = c
	}

	disks, err := r.computeDisks(ctx)
	if err != nil {
		return nil, err
	}
	nics, err := r.computeNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	var networkInfo *handler.NetworkInfo
	if len(nics) > 0 {
		networkInfo = nics[0].AssignedNetwork
	}
	editParameter := r.computeEditParameter(ctx, networkInfo)

	return &computedServerGroupInstance{
		instruction:       r.instruction,
		setupGracePeriod:  r.setupGracePeriod,
		server:            r.server,
		zone:              r.zone,
		disks:             disks,
		diskEditParameter: editParameter,
		cloudConfig:       r.def.Template.CloudConfig.String(),
		networkInterfaces: nics,
		shutdownForce:     r.def.ShutdownForce,
		parent:            parentComputed,
	}, nil
}

func (r *ResourceServerGroupInstance) refresh(ctx *RequestContext) error {
	if r.instruction == handler.ResourceInstructions_DELETE {
		return nil
	}

	serverOp := iaas.NewServerOp(r.apiClient)
	// サーバが存在したらIDで検索する
	if !r.server.ID.IsEmpty() {
		server, err := serverOp.Read(ctx, r.zone, r.server.ID)
		if err != nil {
			return err
		}
		r.server = server
		return nil
	}

	// 存在しなかった(新規作成)の場合はセレクタから検索
	selector := &ResourceSelector{Names: []string{r.server.Name}}
	found, err := serverOp.Find(ctx, r.zone, selector.findCondition())
	if err != nil {
		return err
	}
	if len(found.Servers) == 0 {
		return fmt.Errorf("server not found with: zone=%s Filter='%s'", r.zone, selector.String())
	}
	if len(found.Servers) > 1 {
		return fmt.Errorf("invalid state: found multiple server with: zone=%s Filter='%s'", r.zone, selector.String())
	}
	r.server = found.Servers[0]
	return nil
}

func (r *ResourceServerGroupInstance) computeEditParameter(ctx *RequestContext, networkInfo *handler.NetworkInfo) *handler.ServerGroupInstance_EditParameter {
	if r.instruction != handler.ResourceInstructions_CREATE || r.def.Template.EditParameter == nil {
		return nil
	}

	tmpl := r.def.Template.EditParameter

	if tmpl.Disabled {
		return nil
	}

	if networkInfo == nil {
		networkInfo = &handler.NetworkInfo{}
	}

	var sshKeys []string
	for _, key := range tmpl.SSHKeys {
		sshKeys = append(sshKeys, key.String())
	}

	var startupScripts []string
	for _, ss := range tmpl.StartupScripts {
		startupScripts = append(startupScripts, ss.String())
	}

	return &handler.ServerGroupInstance_EditParameter{
		HostName:            tmpl.HostName(r.server.Name, r.indexInGroup),
		Password:            tmpl.Password,
		DisablePasswordAuth: tmpl.DisablePasswordAuth,
		EnableDhcp:          tmpl.EnableDHCP,
		ChangePartitionUuid: tmpl.ChangePartitionUUID,
		SshKeys:             sshKeys,
		StartupScripts:      startupScripts, // Note: この段階ではGoテンプレートは未評価のまま渡す。

		// これらは必要に応じてHandlerが設定する
		IpAddress:      networkInfo.IpAddress,
		NetworkMaskLen: networkInfo.Netmask,
		DefaultRoute:   networkInfo.Gateway,
	}
}

func (r *ResourceServerGroupInstance) computeDisks(ctx *RequestContext) ([]*handler.ServerGroupInstance_Disk, error) {
	var disks []*handler.ServerGroupInstance_Disk

	if r.instruction != handler.ResourceInstructions_CREATE {
		// 既存ディスクの情報を反映する
		for _, disk := range r.server.Disks {
			// *iaas.Server#Disksから参照できない項目は空のままとしておく(必要に応じてハンドラ側でAPIを叩く)
			disks = append(disks, &handler.ServerGroupInstance_Disk{
				Id:              disk.ID.String(),
				Zone:            r.zone,
				SourceArchiveId: "",
				SourceDiskId:    "",
				Plan:            types.DiskPlanNameMap[disk.DiskPlanID],
				Connection:      disk.Connection.String(),
				Size:            uint32(disk.GetSizeGB()),
				Name:            disk.Name,
				Tags:            []string{},
				Description:     "",
				IconId:          "",
			})
		}
		return disks, nil
	}

	// 新規作成時は必要な値を計算して渡す
	for i, tmpl := range r.def.Template.Disks {
		sourceArchiveID, sourceDiskID, err := tmpl.FindDiskSource(ctx, r.apiClient, r.zone)
		if err != nil {
			return nil, err
		}
		iconId, err := tmpl.FindIconID(ctx, r.apiClient)
		if err != nil {
			return nil, err
		}

		disks = append(disks, &handler.ServerGroupInstance_Disk{
			Id:              "",
			Zone:            r.zone,
			SourceArchiveId: sourceArchiveID,
			SourceDiskId:    sourceDiskID,
			Plan:            tmpl.Plan,
			Connection:      tmpl.Connection,
			Size:            uint32(tmpl.Size),
			Name:            tmpl.DiskName(r.server.Name, i),
			Tags:            tmpl.Tags,
			Description:     tmpl.Description,
			IconId:          iconId,
		})
	}
	return disks, nil
}

func (r *ResourceServerGroupInstance) computeNetworkInterfaces(ctx *RequestContext) ([]*handler.ServerGroupInstance_NIC, error) {
	var nics []*handler.ServerGroupInstance_NIC
	if r.instruction != handler.ResourceInstructions_CREATE {
		for i, nic := range r.server.Interfaces {
			upstream := nic.SwitchID.String()
			if nic.UpstreamType == types.UpstreamNetworkTypes.Shared {
				upstream = "shared"
			}

			var exposeInfo *handler.ServerGroupInstance_ExposeInfo
			if len(r.def.Template.NetworkInterfaces) > i {
				info := r.def.Template.NetworkInterfaces[i].ExposeInfo
				if info != nil {
					exposeInfo = info.ToExposeInfo()
				}
			}

			nics = append(nics, &handler.ServerGroupInstance_NIC{
				Upstream:        upstream,
				PacketFilterId:  nic.PacketFilterID.String(),
				UserIpAddress:   nic.UserIPAddress,
				AssignedNetwork: assignedNetwork(nic, i),
				ExposeInfo:      exposeInfo,
			})
		}
		return nics, nil
	}

	for i, tmpl := range r.def.Template.NetworkInterfaces {
		upstream, err := r.findNetworkUpstream(ctx, tmpl.Upstream)
		if err != nil {
			return nil, err
		}

		var exposeInfo *handler.ServerGroupInstance_ExposeInfo
		if len(r.def.Template.NetworkInterfaces) > i {
			info := r.def.Template.NetworkInterfaces[i].ExposeInfo
			if info != nil {
				exposeInfo = info.ToExposeInfo()
			}
		}

		packetFilterId, err := tmpl.FindPacketFilterId(ctx, r.apiClient, r.zone)
		if err != nil {
			return nil, err
		}
		nic := &handler.ServerGroupInstance_NIC{
			Upstream:       upstream,
			PacketFilterId: packetFilterId,
			ExposeInfo:     exposeInfo,
		}

		if upstream != "shared" {
			ip, mask, err := tmpl.IPAddressByIndexFromCidrBlock(r.indexInGroup)
			if err != nil {
				return nil, err
			}
			nic.UserIpAddress = ip
			nic.AssignedNetwork = &handler.NetworkInfo{
				IpAddress: ip,
				Netmask:   uint32(mask),
				Gateway:   tmpl.DefaultRoute,
				Index:     uint32(i),
			}
		}
		nics = append(nics, nic)
	}
	return nics, nil
}

func (r *ResourceServerGroupInstance) findNetworkUpstream(ctx *RequestContext, upstream *ServerGroupNICUpstream) (string, error) {
	if upstream == nil || upstream.UpstreamShared() {
		return "shared", nil
	}
	selector := upstream.Selector()
	if selector == nil {
		return "", fmt.Errorf("network interface: upstream has invalid value: %#+v", upstream)
	}

	found, err := iaas.NewSwitchOp(r.apiClient).Find(ctx, r.zone, selector.findCondition())
	if err != nil {
		return "", err
	}
	if len(found.Switches) == 0 {
		return "", fmt.Errorf("network interface: upstream not found with: {zone:%s, %v}", r.zone, selector)
	}
	if len(found.Switches) > 1 {
		return "", fmt.Errorf("multiple source archive found with: {zone: %s, %v}, switches: %v", r.zone, selector, found.Switches)
	}
	return found.Switches[0].ID.String(), nil
}

func (r *ResourceServerGroupInstance) isHealthy(ctx *RequestContext) (bool, error) {
	if !r.server.InstanceStatus.IsUp() {
		return false, nil
	}

	if r.parent != nil {
		requests, err := r.healthCheckRequests(ctx)
		if err != nil {
			return false, err
		}
		return r.parent.(*ParentResource).IsChildResourceHealthy(ctx, requests)
	}

	return true, nil
}

func (r *ResourceServerGroupInstance) healthCheckRequests(ctx *RequestContext) ([]*ChildResourceHealthCheckRequest, error) {
	nics, err := r.computeNetworkInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	var requests []*ChildResourceHealthCheckRequest
	for _, nic := range nics {
		if nic.ExposeInfo == nil {
			continue
		}

		// VIPsやPortsが指定されていない場合、1回はループを回すためにダミーを投入しておく
		vips := nic.ExposeInfo.Vips
		if len(vips) == 0 {
			vips = append(vips, "")
		}

		ports := nic.ExposeInfo.Ports
		if len(ports) == 0 {
			ports = append(ports, 0)
		}

		for _, port := range ports {
			for _, vip := range vips {
				ip := nic.GetUserIpAddress()
				if ip == "" {
					ip = nic.AssignedNetwork.IpAddress
				}

				requests = append(requests, &ChildResourceHealthCheckRequest{
					VIP:       vip,
					IPAddress: ip,
					Port:      int(port),
				})
			}
		}
	}
	return requests, nil
}
