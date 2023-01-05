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

package server

import (
	"bytes"
	"text/template"
	"time"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/types"
	diskBuilder "github.com/sacloud/iaas-service-go/disk/builder"
	serverBuilder "github.com/sacloud/iaas-service-go/server/builder"
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
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)

	server := req.Desired.GetServerGroupInstance()
	if server != nil {
		switch req.Instruction {
		case handler.ResourceInstructions_CREATE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.createServer(ctx, req, server)
		case handler.ResourceInstructions_DELETE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.deleteServer(ctx, req, server)
		}
	}

	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *HorizontalScaleHandler) createServer(ctx *handlers.HandlerContext, req *handler.HandleRequest, server *handler.ServerGroupInstance) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
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

	if err := ctx.Report(handler.HandleResponse_RUNNING, "creating..."); err != nil {
		return err
	}

	created, err := sb.Build(ctx, server.Zone)
	if err != nil {
		return err
	}

	serverOp := iaas.NewServerOp(h.APICaller())
	createdServer, err := serverOp.Read(ctx, server.Zone, created.ServerID)
	if err != nil {
		return err
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"created: {Zone:%s, ID:%s, Name:%s}", createdServer.Zone.Name, createdServer.ID, createdServer.Name); err != nil {
		return err
	}

	diskOp := iaas.NewDiskOp(h.APICaller())
	for i, d := range server.Disks {
		editParameter, err := h.diskEditParameter(server, i)
		if err != nil {
			return err
		}
		builder := (&diskBuilder.Director{
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

		if err := ctx.Report(handler.HandleResponse_RUNNING, "creating disk[%d]...", i); err != nil {
			return err
		}

		buildResult, err := builder.Build(ctx, server.Zone, createdServer.ID)
		if err != nil {
			return err
		}
		disk, err := diskOp.Read(ctx, server.Zone, buildResult.DiskID)
		if err != nil {
			return err
		}

		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"created disk[%d]: {Zone:%s, ID:%s, Name:%s, ServerID:%s}",
			i, server.Zone, disk.ID, disk.Name, createdServer.ID); err != nil {
			return err
		}
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING, "starting..."); err != nil {
		return err
	}
	if err := power.BootServer(ctx, serverOp, server.Zone, createdServer.ID, server.CloudConfig); err != nil {
		return err
	}

	if req.SetupGracePeriod > 0 {
		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"waiting for setup to complete: setup_grace_period=%d", req.SetupGracePeriod); err != nil {
			return err
		}
		time.Sleep(time.Duration(req.SetupGracePeriod) * time.Second)
	}

	return ctx.Report(handler.HandleResponse_DONE, "started")
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

func (h *HorizontalScaleHandler) deleteServer(ctx *handlers.HandlerContext, req *handler.HandleRequest, server *handler.ServerGroupInstance) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	serverOp := iaas.NewServerOp(h.APICaller())
	current, err := serverOp.Read(ctx, server.Zone, types.StringID(server.Id))
	if err != nil {
		return err
	}

	if current.InstanceStatus.IsUp() {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "shutting down..."); err != nil {
			return err
		}

		if err := power.ShutdownServer(ctx, serverOp, server.Zone, current.ID, server.ShutdownForce); err != nil {
			return err
		}

		if err := ctx.Report(handler.HandleResponse_RUNNING, "shut down"); err != nil {
			return err
		}
	}

	var diskIDs []types.ID
	for _, disk := range current.Disks {
		diskIDs = append(diskIDs, disk.ID)
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING, "deleting...: {Disks:%s}", diskIDs); err != nil {
		return err
	}

	if err := serverOp.DeleteWithDisks(ctx, server.Zone, current.ID, &iaas.ServerDeleteWithDisksRequest{IDs: diskIDs}); err != nil {
		return err
	}

	return ctx.Report(handler.HandleResponse_DONE, "deleted: {Zone:%s, ID:%s, Name:%s, Disks:%s}", server.Zone, server.Id, server.Name, diskIDs)
}
