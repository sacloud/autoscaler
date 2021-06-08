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

package elb

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type VerticalScaleHandler struct {
	handlers.SakuraCloudFlagCustomizer
	Logger *log.Logger
}

func (h *VerticalScaleHandler) Name() string {
	return "elb-vertical-scaler"
}

func (h *VerticalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *VerticalScaleHandler) GetLogger() *log.Logger {
	return h.Logger
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

	elb := req.Desired.GetElb()
	if elb != nil && req.Instruction == handler.ResourceInstructions_UPDATE {
		// TODO 入力値のバリデーション
		if err := h.handleELB(ctx, req, elb, sender); err != nil {
			return err
		}
	}

	return nil
}

func (h *VerticalScaleHandler) handleELB(ctx context.Context, req *handler.HandleRequest, elb *handler.ELB, sender handlers.ResponseSender) error {
	elbOp := sacloud.NewProxyLBOp(h.APIClient())

	_, err := elbOp.ChangePlan(ctx, types.StringID(elb.Id), &sacloud.ProxyLBChangePlanRequest{
		ServiceClass: types.ProxyLBServiceClass(types.EProxyLBPlan(elb.Plan), types.EProxyLBRegion(elb.Region)),
	})
	if err != nil {
		return err
	}
	return sender.Send(&handler.HandleResponse{
		ScalingJobId: req.ScalingJobId,
		Status:       handler.HandleResponse_DONE,
	})
}
