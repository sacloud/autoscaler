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

package router

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
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
	return "router-vertical-scaler"
}

func (h *VerticalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *VerticalScaleHandler) Handle(req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := context.TODO()

	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_ACCEPTED,
	}); err != nil {
		return err
	}

	router := req.Desired.GetRouter()
	if router != nil && req.Instruction == handler.ResourceInstructions_UPDATE {
		if err := h.handleRouter(ctx, req, router, sender); err != nil {
			return err
		}
	} else {
		return sender.Send(&handler.HandleResponse{
			ScalingJobId: req.ScalingJobId,
			Status:       handler.HandleResponse_IGNORED,
		})
	}

	return nil
}

func (h *VerticalScaleHandler) handleRouter(ctx context.Context, req *handler.HandleRequest, router *handler.Router, sender handlers.ResponseSender) error {
	if err := sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_RUNNING,
		Log:          fmt.Sprintf("router plan changing - to %d Mbps", router.BandWidth),
	}); err != nil {
		return err
	}

	routerOp := sacloud.NewInternetOp(h.APICaller())

	updated, err := routerOp.UpdateBandWidth(ctx, router.Zone, types.StringID(router.Id), &sacloud.InternetUpdateBandWidthRequest{
		BandWidthMbps: int(router.BandWidth),
	})
	if err != nil {
		return err
	}
	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
		Log:          fmt.Sprintf("router plan changed - resource ID changed: from %s to %s", router.Id, updated.ID.String()),
	})
}
