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
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type VerticalScaleHandler struct {
	handlers.SakuraCloudFlagCustomizer
}

func (h *VerticalScaleHandler) Name() string {
	return "server-vertical-scaler"
}

func (h *VerticalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *VerticalScaleHandler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := context.TODO()

	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
		Log:          fmt.Sprintf("%s: accepted: %s", h.Name(), req.String()),
	}); err != nil {
		return err
	}

	server := req.Desired.GetServer()
	if server != nil && server.Instruction == handler.ResourceInstructions_UPDATE {
		// TODO 入力値のバリデーション
		if err := h.handleServer(ctx, req, server, sender); err != nil {
			return err
		}
	}

	return nil
}

func (h *VerticalScaleHandler) handleServer(ctx context.Context, req *handler.HandleRequest, server *handler.Server, sender handlers.ResponseSender) error {
	serverOp := sacloud.NewServerOp(h.APIClient())

	// TODO サーバが起動していた場合はシャットダウン

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
	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_RUNNING,
		Log:          fmt.Sprintf("server plan changed - resource ID cahnged: from %s to %s", server.Id, updated.ID.String()),
	})
}
