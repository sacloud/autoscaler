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

package server

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	diskBuilder "github.com/sacloud/libsacloud/v2/helper/builder/disk"
	serverBuilder "github.com/sacloud/libsacloud/v2/helper/builder/server"
	"github.com/sacloud/libsacloud/v2/helper/power"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/ostype"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type HorizontalScaleHandler struct {
	handlers.HandlerLogger
	*builtins.SakuraCloudAPIClient
}

func NewHorizontalScaleHandler() *HorizontalScaleHandler {
	return &HorizontalScaleHandler{
		SakuraCloudAPIClient: &builtins.SakuraCloudAPIClient{},
	}
}

func (h *HorizontalScaleHandler) Name() string {
	return "server-horizontal-scaler"
}

func (h *HorizontalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *HorizontalScaleHandler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := context.Background()
	server := req.Desired.GetServerGroupInstance()

	switch req.Instruction {
	case handler.ResourceInstructions_CREATE:
		return h.createServer(ctx, req, server, sender)
	case handler.ResourceInstructions_DELETE:
		return h.deleteServer(ctx, req, server, sender)
	default:
		return sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_IGNORED,
		})
	}
}

func (h *HorizontalScaleHandler) createServer(ctx context.Context, req *handler.HandleRequest, server *handler.ServerGroupInstance, sender handlers.ResponseSender) error {
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_RUNNING,
	}); err != nil {
		return err
	}

	commitment := types.Commitments.Standard
	if server.DedicatedCpu {
		commitment = types.Commitments.DedicatedCPU
	}
	sb := serverBuilder.Builder{
		Name:            server.Name,
		CPU:             int(server.Core),
		MemoryGB:        int(server.Memory),
		Commitment:      commitment,
		Generation:      types.PlanGenerations.Default,
		InterfaceDriver: types.InterfaceDriverMap[server.InterfaceDriver],
		Description:     server.Description,
		IconID:          types.StringID(server.IconId),
		Tags:            server.Tags,
		BootAfterCreate: false,
		CDROMID:         types.StringID(server.CdRomId),
		PrivateHostID:   types.StringID(server.PrivateHostId),
		NIC:             h.networkInterface(server),
		AdditionalNICs:  h.additionalNetworkInterfaces(server),
		Client:          serverBuilder.NewBuildersAPIClient(h.APICaller()),
	}

	created, err := sb.Build(ctx, server.Zone)
	if err != nil {
		return err
	}

	serverOp := sacloud.NewServerOp(h.APICaller())
	createdServer, err := serverOp.Read(ctx, server.Zone, created.ServerID)
	if err != nil {
		return err
	}

	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_RUNNING,
		Log:          fmt.Sprintf("server created: ID:%s, Name:%s", createdServer.ID, createdServer.Name),
	}); err != nil {
		return err
	}

	diskOp := sacloud.NewDiskOp(h.APICaller())
	//var createdDisks []*sacloud.Disk
	for i, d := range server.Disks {
		editParameter, err := h.diskEditParameter(server, i)
		if err != nil {
			return err
		}
		builder := (&diskBuilder.Director{
			OSType:          ostype.StrToOSType(d.OsType),
			Name:            d.Name,
			SizeGB:          int(d.Size),
			DistantFrom:     nil, // TODO .protoに追記
			PlanID:          types.DiskPlanIDMap[d.Plan],
			Connection:      types.DiskConnectionMap[d.Connection],
			Description:     d.Description,
			Tags:            d.Tags,
			IconID:          types.StringID(d.IconId),
			SourceDiskID:    types.StringID(d.SourceDiskId),
			SourceArchiveID: types.StringID(d.SourceArchiveId),
			EditParameter:   editParameter,
			NoWait:          false,
			Client:          diskBuilder.NewBuildersAPIClient(h.APICaller()),
		}).Builder()

		buildResult, err := builder.Build(ctx, server.Zone, createdServer.ID)
		if err != nil {
			return err
		}
		disk, err := diskOp.Read(ctx, server.Zone, buildResult.DiskID)
		if err != nil {
			return err
		}

		//createdDisks = append(createdDisks, disk)
		if err := sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_RUNNING,
			Log:          fmt.Sprintf("disk created: ServerID:%s, ID:%s, Name:%s", createdServer.ID, disk.ID, disk.Name),
		}); err != nil {
			return err
		}
	}

	if err := power.BootServer(ctx, serverOp, server.Zone, createdServer.ID); err != nil {
		return err
	}

	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
		Log:          fmt.Sprintf("server created: ID:%s, Name:%s", createdServer.ID, createdServer.Name),
	})
}

func (h *HorizontalScaleHandler) diskEditParameter(server *handler.ServerGroupInstance, diskIndex int) (*diskBuilder.EditRequest, error) {
	if diskIndex != 0 || server.EditParameter == nil {
		return nil, nil
	}

	var sshKeyIDs []types.ID
	for _, keyID := range server.EditParameter.SshKeyIds {
		id := types.StringID(keyID)
		if !id.IsEmpty() {
			sshKeyIDs = append(sshKeyIDs, id)
		}
	}

	startupScripts, err := h.execStartupScriptTemplate(server, server.EditParameter.StartupScripts)
	if err != nil {
		return nil, err
	}

	return &diskBuilder.EditRequest{
		HostName:            server.EditParameter.HostName,
		Password:            server.EditParameter.Password,
		DisablePWAuth:       server.EditParameter.DisablePasswordAuth,
		EnableDHCP:          server.EditParameter.EnableDhcp,
		ChangePartitionUUID: server.EditParameter.ChangePartitionUuid,
		IPAddress:           server.EditParameter.IpAddress,
		NetworkMaskLen:      int(server.EditParameter.NetworkMaskLen),
		DefaultRoute:        server.EditParameter.DefaultRoute,
		SSHKeys:             server.EditParameter.SshKeys,
		SSHKeyIDs:           sshKeyIDs,
		IsSSHKeysEphemeral:  false,
		IsNotesEphemeral:    true,
		NoteContents:        startupScripts,
		Notes:               nil,
	}, nil
}

func (h *HorizontalScaleHandler) execStartupScriptTemplate(server *handler.ServerGroupInstance, scripts []string) ([]string, error) {
	var startupScripts []string
	for _, ss := range scripts {
		tmpl, err := template.New(server.Name).Parse(ss)
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBufferString("")
		if err := tmpl.Execute(buf, server); err != nil {
			return nil, err
		}
		startupScripts = append(startupScripts, buf.String())
	}
	return startupScripts, nil
}

func (h *HorizontalScaleHandler) networkInterface(server *handler.ServerGroupInstance) serverBuilder.NICSettingHolder {
	if len(server.NetworkInterfaces) == 0 {
		return nil
	}
	nic := server.NetworkInterfaces[0]
	if nic.Upstream == "shared" {
		return &serverBuilder.SharedNICSetting{PacketFilterID: types.StringID(nic.PacketFilterId)}
	}

	return &serverBuilder.ConnectedNICSetting{
		SwitchID:         types.StringID(nic.Upstream),
		DisplayIPAddress: nic.UserIpAddress,
		PacketFilterID:   types.StringID(nic.PacketFilterId),
	}
}

func (h *HorizontalScaleHandler) additionalNetworkInterfaces(server *handler.ServerGroupInstance) []serverBuilder.AdditionalNICSettingHolder {
	if len(server.NetworkInterfaces) < 1 {
		return nil
	}

	var nics []serverBuilder.AdditionalNICSettingHolder
	for _, nic := range server.NetworkInterfaces[1:] {
		nics = append(nics, &serverBuilder.ConnectedNICSetting{
			SwitchID:         types.StringID(nic.Upstream),
			DisplayIPAddress: nic.UserIpAddress,
			PacketFilterID:   types.StringID(nic.PacketFilterId),
		})
	}
	return nics
}

func (h *HorizontalScaleHandler) deleteServer(ctx context.Context, req *handler.HandleRequest, server *handler.ServerGroupInstance, sender handlers.ResponseSender) error {
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_RUNNING,
	}); err != nil {
		return err
	}

	serverOp := sacloud.NewServerOp(h.APICaller())
	current, err := serverOp.Read(ctx, server.Zone, types.StringID(server.Id))
	if err != nil {
		return err
	}

	if current.InstanceStatus.IsUp() {
		// シャットダウン
		if err := power.ShutdownServer(ctx, serverOp, server.Zone, current.ID, server.ShutdownForce); err != nil {
			return err
		}
		if err := sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_RUNNING,
			Log:          fmt.Sprintf("shutting down server: %s", server.Id),
		}); err != nil {
			return err
		}
	}

	var diskIDs []types.ID
	for _, disk := range current.Disks {
		diskIDs = append(diskIDs, disk.ID)
	}
	if err := serverOp.DeleteWithDisks(ctx, server.Zone, current.ID, &sacloud.ServerDeleteWithDisksRequest{IDs: diskIDs}); err != nil {
		return err
	}

	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
		Log:          fmt.Sprintf("server deleted: ID:%s, Name:%s, Disk:%s", current.ID, current.Name, diskIDs),
	})
}