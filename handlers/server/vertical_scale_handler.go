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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/helper/power"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type VerticalScaleHandler struct {
	handlers.HandlerLogger
	*builtins.SakuraCloudAPIClient
}

func NewVerticalScaleHandler() *VerticalScaleHandler {
	return &VerticalScaleHandler{
		SakuraCloudAPIClient: &builtins.SakuraCloudAPIClient{},
	}
}

func (h *VerticalScaleHandler) Name() string {
	return "server-vertical-scaler"
}

func (h *VerticalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *VerticalScaleHandler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(req.ScalingJobId, sender)
	server := req.Desired.GetServer()

	if server != nil {
		switch req.Instruction {
		case handler.ResourceInstructions_UPDATE:
			if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
				return err
			}
			return h.handleServer(ctx, req, server)
		}
	}
	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *VerticalScaleHandler) handleServer(ctx *handlers.HandlerContext, req *handler.HandleRequest, server *handler.Server) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING); err != nil {
		return err
	}

	serverOp := sacloud.NewServerOp(h.APICaller())

	current, err := serverOp.Read(ctx, server.Zone, types.StringID(server.Id))
	if err != nil {
		return err
	}

	shouldReboot := false
	if current.InstanceStatus.IsUp() {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "shutting down..."); err != nil {
			return err
		}

		if err := power.ShutdownServer(ctx, serverOp, server.Zone, types.StringID(server.Id), server.ShutdownForce); err != nil {
			return err
		}

		if err := ctx.Report(handler.HandleResponse_RUNNING, "shut down"); err != nil {
			return err
		}
		shouldReboot = true
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"plan changing: {Core:%d, Memory:%d}", server.Core, server.Memory); err != nil {
		return err
	}

	commitment := types.Commitments.Standard
	if server.DedicatedCpu {
		commitment = types.Commitments.DedicatedCPU
	}
	updated, err := serverOp.ChangePlan(ctx, server.Zone, types.StringID(server.Id), &sacloud.ServerChangePlanRequest{
		CPU:                  int(server.Core),
		MemoryMB:             int(server.Memory * size.GiB),
		ServerPlanGeneration: types.PlanGenerations.Default, // TODO プランの世代はどう指定するか?
		ServerPlanCommitment: commitment,
	})
	if err != nil {
		return err
	}

	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"plan changed: {ServerIDFrom:%s, ServerIDTo:%s}", server.Id, updated.ID); err != nil {
		return err
	}

	if shouldReboot {
		if err := ctx.Report(handler.HandleResponse_RUNNING, "starting..."); err != nil {
			return err
		}

		if err := power.BootServer(ctx, serverOp, server.Zone, updated.ID); err != nil {
			return err
		}

		if err := ctx.Report(handler.HandleResponse_RUNNING, "started"); err != nil {
			return err
		}
	}

	return ctx.Report(handler.HandleResponse_DONE)
}
