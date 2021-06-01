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

package core

import (
	"context"
	"testing"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/stretchr/testify/require"
)

func testServer() *Server {
	return &Server{
		ResourceBase: &ResourceBase{
			TypeName: "Server",
			TargetSelector: &ResourceSelector{
				Names: []string{"test-server"},
				Zones: testZones,
			},
		},
		Zone: testZone,
		Plans: []ServerPlan{
			{Core: 1, Memory: 1},
			{Core: 2, Memory: 4},
			{Core: 4, Memory: 8},
		},
	}
}

func testContext() *Context {
	return NewContext(context.Background(), &requestInfo{
		requestType:       requestTypeUp,
		source:            "default",
		action:            "default",
		resourceGroupName: "web",
	})
}

func initTestServer(t *testing.T) func() {
	serverOp := sacloud.NewServerOp(testAPIClient)
	server, err := serverOp.Create(context.Background(), testZone, &sacloud.ServerCreateRequest{
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

	return func() {
		if err := serverOp.Delete(context.Background(), testZone, server.ID); err != nil {
			t.Logf("[WARN] deleting server failed: %s", err)
		}
	}
}

func TestServer_Validate(t *testing.T) {
	defer initTestServer(t)()

	t.Run("returns error if selector is empty", func(t *testing.T) {
		empty := &Server{
			ResourceBase: &ResourceBase{TypeName: "Server"},
		}
		err := empty.Validate()
		require.Error(t, err)
		require.EqualError(t, err, "selector: required")
	})

	t.Run("returns error if selector.Zones is empty", func(t *testing.T) {
		empty := &Server{
			ResourceBase: &ResourceBase{
				TypeName:       "Server",
				TargetSelector: &ResourceSelector{},
			},
		}
		err := empty.Validate()
		require.Error(t, err)
		require.EqualError(t, err, "selector.Zones: least one value required")
	})
}

func TestServer_Computed(t *testing.T) {
	defer initTestServer(t)()

	ctx := testContext()

	t.Run("returns error if selector has invalid value", func(t *testing.T) {
		notFound := &Server{
			ResourceBase: &ResourceBase{
				TypeName: "Server",
				TargetSelector: &ResourceSelector{
					ID:    123456789012,
					Zones: testZones,
				},
			},
		}

		_, err := notFound.Compute(ctx, testAPIClient)
		require.Error(t, err)
	})

	t.Run("returns UPDATE instruction if selector has valid value", func(t *testing.T) {
		running := &Server{
			ResourceBase: &ResourceBase{
				TypeName: "Server",
				TargetSelector: &ResourceSelector{
					Names: []string{"test-server"},
					Zones: testZones,
				},
			},
			Plans: []ServerPlan{
				{Core: 1, Memory: 1},
				{Core: 2, Memory: 4},
				{Core: 4, Memory: 8},
			},
		}

		computed, err := running.Compute(ctx, testAPIClient)
		require.NoError(t, err)
		require.NotNil(t, computed)
		require.Len(t, computed, 1)
		require.Equal(t, handler.ResourceInstructions_UPDATE, computed[0].Instruction())

		current := computed[0].Current()
		require.NotNil(t, current)

		desired := computed[0].Desired()
		require.NotNil(t, desired)
	})

	t.Run("returns desired state that can convert to the request parameter", func(t *testing.T) {
		ctx := testContext()
		server := testServer()
		computed, err := server.Compute(ctx, testAPIClient)
		require.NoError(t, err)
		require.Len(t, computed, 1)

		handlerReq := computed[0].Desired()
		require.NotNil(t, handlerReq)

		desiredServer := handlerReq.GetServer()
		require.NotNil(t, desiredServer)

		// Server.Plansで指定した次のプランが返されるはず
		require.Equal(t, uint32(4), desiredServer.Core)
		require.Equal(t, uint32(8), desiredServer.Memory)
		require.Equal(t, server.DedicatedCPU, desiredServer.DedicatedCpu)
	})

	t.Run("stores results to own cache", func(t *testing.T) {
		ctx := testContext()
		server := testServer()
		computed, err := server.Compute(ctx, testAPIClient)
		require.NoError(t, err)
		require.Len(t, computed, 1)

		cached := server.Computed()
		require.Len(t, cached, 1)
		require.Equal(t, computed, cached)

		server.ClearCache()
		cached = server.Computed()
		require.Len(t, cached, 0)
	})
}
