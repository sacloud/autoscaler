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

package elb

import (
	"context"
	"time"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/iaas-api-go/helper/plans"
	"github.com/sacloud/iaas-api-go/types"
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
	return "elb-vertical-scaler"
}

func (h *VerticalScaleHandler) Version() string {
	return version.FullVersion()
}

func (h *VerticalScaleHandler) Handle(parentCtx context.Context, req *handler.HandleRequest, sender handlers.ResponseSender) error {
	ctx := handlers.NewHandlerContext(parentCtx, req.ScalingJobId, sender)

	elb := req.Desired.GetElb()
	if elb != nil && req.Instruction == handler.ResourceInstructions_UPDATE {
		if err := ctx.Report(handler.HandleResponse_ACCEPTED); err != nil {
			return err
		}

		return h.handleELB(ctx, req, elb)
	}
	return ctx.Report(handler.HandleResponse_IGNORED)
}

func (h *VerticalScaleHandler) handleELB(ctx *handlers.HandlerContext, req *handler.HandleRequest, elb *handler.ELB) error {
	if err := ctx.Report(handler.HandleResponse_RUNNING,
		"plan changing...: {Desired CPS:%d}", elb.Plan); err != nil {
		return err
	}

	updated, err := plans.ChangeProxyLBPlan(ctx, h.APICaller(), types.StringID(elb.Id), int(elb.Plan))
	if err != nil {
		return err
	}
	if req.SetupGracePeriod > 0 {
		if err := ctx.Report(handler.HandleResponse_RUNNING,
			"waiting for setup to complete: setup_grace_period=%d", req.SetupGracePeriod); err != nil {
			return err
		}
		time.Sleep(time.Duration(req.SetupGracePeriod) * time.Second)
	}

	return ctx.Report(
		handler.HandleResponse_DONE,
		"plan changed: {ID:%s, CPS:%d}", updated.ID, elb.Plan,
	)
}
