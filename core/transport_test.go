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
	"context"
	"strings"
	"testing"

	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/request"
	"github.com/stretchr/testify/require"
)

var _ request.ScalingServiceServer = (*fakeScalingService)(nil)

type fakeScalingService struct {
	request.UnimplementedScalingServiceServer
}

func (s *fakeScalingService) Up(context.Context, *request.ScalingRequest) (*request.ScalingResponse, error) {
	return &request.ScalingResponse{
		ScalingJobId: "default",
		Status:       request.ScalingJobStatus_JOB_DONE,
	}, nil
}

func (s *fakeScalingService) Down(context.Context, *request.ScalingRequest) (*request.ScalingResponse, error) {
	return &request.ScalingResponse{
		ScalingJobId: "default",
		Status:       request.ScalingJobStatus_JOB_DONE,
	}, nil
}

func TestTransport(t *testing.T) {
	tests := []struct {
		name       string
		listenAddr string
		wantError  bool
	}{
		{
			name:       "unix domain socket",
			listenAddr: "unix:autoscaler.sock",
			wantError:  false,
		},
		{
			name:       "h2c",
			listenAddr: "localhost:0",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, listener, cleanup, err := grpcutil.Server(&grpcutil.ListenerOption{
				Address: tt.listenAddr,
			})
			if err != nil {
				t.Fatal(err)
			}
			srv := &fakeScalingService{}
			request.RegisterScalingServiceServer(server, srv)

			defer func() {
				server.GracefulStop()
				cleanup()
			}()

			errCh := make(chan error)
			go func() {
				if err := server.Serve(listener); err != nil {
					errCh <- err
				}
			}()

			addr := tt.listenAddr
			if !strings.HasPrefix(addr, "unix:") {
				addr = listener.Addr().String()
			}
			opts := &grpcutil.DialOption{Destination: addr}
			conn, cleanup2, err := grpcutil.DialContext(context.Background(), opts)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup2()

			req := request.NewScalingServiceClient(conn)
			_, err = req.Up(context.Background(), &request.ScalingRequest{})
			require.Equal(t, err != nil, tt.wantError, "unexpected error: expected: %t, actual:%s", tt.wantError, err)
			if err != nil {
				t.Logf("name: %s, error: %s", tt.name, err)
			}
		})
	}
}
