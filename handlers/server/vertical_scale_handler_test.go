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
	"io"
	"testing"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type fakeSender struct {
	buf *bytes.Buffer
}

func (s *fakeSender) Send(res *handler.HandleResponse) error {
	_, err := io.Copy(s.buf, bytes.NewBufferString(res.Log))
	return err
}

func TestHandler_Handle(t *testing.T) {
	server, cleanup := initTestServer(t)
	defer cleanup()

	sender := &fakeSender{buf: bytes.NewBufferString("")}

	type args struct {
		req    *handler.HandleRequest
		sender handlers.ResponseSender
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "scale up",
			args: args{
				req: &handler.HandleRequest{
					Source:       "default",
					ResourceName: "default",
					ScalingJobId: "1",
					Instruction:  handler.ResourceInstructions_UPDATE,
					Desired: &handler.Resource{
						Resource: &handler.Resource_Server{
							Server: &handler.Server{
								Id:              server.ID.String(),
								AssignedNetwork: nil,
								Core:            4,
								Memory:          8,
								Zone:            test.Zone,
							},
						},
					},
				},
				sender: sender,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewVerticalScaleHandler()
			h.SetAPICaller(test.APIClient)

			if err := h.Handle(tt.args.req, tt.args.sender); (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func initTestServer(t *testing.T) (*sacloud.Server, func()) {
	serverOp := sacloud.NewServerOp(test.APIClient)
	server, err := serverOp.Create(context.Background(), test.Zone, &sacloud.ServerCreateRequest{
		CPU:                  2,
		MemoryMB:             4 * size.GiB,
		ServerPlanCommitment: types.Commitments.Standard,
		ServerPlanGeneration: types.PlanGenerations.Default,
		ConnectedSwitches:    nil,
		InterfaceDriver:      types.InterfaceDrivers.VirtIO,
		Name:                 "test-server",
	})
	if err != nil {
		t.Fatal(err)
	}

	return server, func() {
		// TODO プラン変更後のサーバのクリーンアップを行いたいが、プラン変更でIDが変わるためここでは行えない。
		// fakeドライバの場合は不要だが以外の場合のも対応したいため、どこかで実装するようにする
	}
}
