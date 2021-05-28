package logging

import (
	"context"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"os"
	"testing"
)

func TestHandler_Handle(t *testing.T) {
	server, cleanup := initTestServer(t)
	defer cleanup()

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
				req:    &handler.HandleRequest{
					Source:            "default",
					Action:            "default",
					ResourceGroupName: "default",
					ScalingJobId:      "1",
					Resources:         []*handler.Resource{
						{Resource: &handler.Resource_Server{
							Server: &handler.Server{
								Status:          handler.ResourceStatus_RUNNING,
								Id:              server.ID.String(),
								AssignedNetwork: nil,
								Core:            4,
								Memory:          8,
							},
						}},
					},
				},
				sender: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			if err := h.Handle(tt.args.req, tt.args.sender); (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var (
	testZone = "is1a"
	testAPIClient = 	api.NewCaller(&api.CallerOptions{
		AccessToken:       "fake",
		AccessTokenSecret: "fake",
		UserAgent:         "sacloud/autoscaler/fake",
		TraceAPI:          os.Getenv("SAKURACLOUD_TRACE") != "",
		TraceHTTP:         os.Getenv("SAKURACLOUD_TRACE") != "",
		FakeMode:          true,
	})
)

func initTestServer(t *testing.T) (*sacloud.Server, func()) {
	serverOp := sacloud.NewServerOp(testAPIClient)
	server, err := serverOp.Create(context.Background(), testZone	, &sacloud.ServerCreateRequest{
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
		if err := serverOp.Delete(context.Background(), testZone, server.ID); err != nil {
			t.Logf("[WARN] deleting server failed: %s", err)
		}
	}
}
