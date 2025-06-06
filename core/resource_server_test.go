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
	"reflect"
	"testing"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/plans"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
	"github.com/stretchr/testify/require"
)

func initTestResourceServer(t *testing.T) *iaas.Server {
	serverOp := iaas.NewServerOp(test.APIClient)
	server, err := serverOp.Create(context.Background(), test.Zone, &iaas.ServerCreateRequest{
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
	return server
}

func TestResourceServer_New_Refresh(t *testing.T) {
	ctx := testContext()
	server := initTestResourceServer(t)

	def := &ResourceDefServer{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "Server",
		},
		Selector: &MultiZoneSelector{
			ResourceSelector: &ResourceSelector{},
		},
		DedicatedCPU:  false,
		Plans:         nil,
		ShutdownForce: false,
	}

	resource, err := NewResourceServer(ctx, test.APIClient, def, nil, test.Zone, server)
	require.NoError(t, err)
	require.NotNil(t, resource)

	serverOp := iaas.NewServerOp(test.APIClient)

	server, err = serverOp.Read(ctx, test.Zone, server.ID)
	require.NoError(t, err)

	// IDを変えるためにプラン変更を実施
	updated, err := plans.ChangeServerPlan(ctx, test.APIClient, test.Zone, server.ID,
		&iaas.ServerChangePlanRequest{
			CPU:                  1,
			MemoryMB:             2 * size.GiB,
			GPU:                  0,
			ServerPlanCPUModel:   "",
			ServerPlanGeneration: types.PlanGenerations.Default,
			ServerPlanCommitment: types.Commitments.Standard,
		},
	)
	require.NoError(t, err)

	// refresh実施
	_, err = resource.Compute(ctx, true)
	require.NoError(t, err)

	require.EqualValues(t, plans.AppendPreviousIDTagIfAbsent(types.Tags{}, server.ID), resource.server.Tags)

	// cleanup
	if err := serverOp.Delete(ctx, test.Zone, updated.ID); err != nil {
		t.Fatal(err)
	}
}

func TestResourceServer_Compute(t *testing.T) {
	server := initTestResourceServer(t)
	defer func() {
		if err := iaas.NewServerOp(test.APIClient).Delete(context.Background(), test.Zone, server.ID); err != nil {
			t.Fatal(err)
		}
	}()
	def := &ResourceDefServer{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "Server",
			DefName:  "default",
		},
		Selector: &MultiZoneSelector{
			ResourceSelector: &ResourceSelector{},
			Zones:            nil,
		},
		DedicatedCPU: false,
		Plans: []*ServerPlan{
			{Core: 1, Memory: 1, Name: "plan1"},
			{Core: 2, Memory: 4, Name: "plan2"},
			{Core: 4, Memory: 8, Name: "plan3"},
		},
		ShutdownForce: false,
	}
	resource := &ResourceServer{
		ResourceBase: &ResourceBase{
			resourceType:     ResourceTypeServer,
			setupGracePeriod: def.SetupGracePeriod(),
		},
		apiClient: test.APIClient,
		server:    server,
		def:       def,
		zone:      test.Zone,
	}

	type args struct {
		ctx     *RequestContext
		refresh bool
	}
	tests := []struct {
		name    string
		args    args
		want    Computed
		wantErr bool
	}{
		{
			name: "up",
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeUp,
					source:           "default",
					resourceName:     "default",
					desiredStateName: "",
				}, test.Logger),
				refresh: false,
			},
			want: &computedServer{
				instruction:      handler.ResourceInstructions_UPDATE,
				server:           server,
				zone:             test.Zone,
				newCPU:           4,
				newMemory:        8,
				setupGracePeriod: defaults.SetupGracePeriods[ResourceTypeServer.String()],
				parent:           nil,
			},
			wantErr: false,
		},
		{
			name: "down",
			args: args{
				ctx: NewRequestContext(context.Background(), &requestInfo{
					requestType:      requestTypeDown,
					source:           "default",
					resourceName:     "default",
					desiredStateName: "",
				}, test.Logger),
				refresh: false,
			},
			want: &computedServer{
				instruction:      handler.ResourceInstructions_UPDATE,
				server:           server,
				zone:             test.Zone,
				newCPU:           1,
				newMemory:        1,
				setupGracePeriod: defaults.SetupGracePeriods[ResourceTypeServer.String()],
				parent:           nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.Compute(tt.args.ctx, tt.args.refresh)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
