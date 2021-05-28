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

func TestServer_Calculate(t *testing.T) {
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

		_, _, err := notFound.Calculate(ctx, testAPIClient)
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

		current, _, err := running.Calculate(ctx, testAPIClient)
		require.NoError(t, err)
		require.NoError(t, err)
		require.NotNil(t, current)
		require.Equal(t, handler.ResourceInstructions_UPDATE, current.Status())
	})

	t.Run("returns scale-upd state", func(t *testing.T) {
		ctx := testContext()
		server := testServer()
		current, desired, err := server.Calculate(ctx, testAPIClient)
		require.NoError(t, err)
		require.NotNil(t, current)
		require.NotNil(t, desired)

		desiredServer, ok := desired.Raw().(*sacloud.Server)
		require.True(t, ok)
		require.NotNil(t, desiredServer)

		// Server.Plansで指定した次のプランが返されるはず
		require.Equal(t, 4, desiredServer.CPU, 2)
		require.Equal(t, 8, size.MiBToGiB(desiredServer.MemoryMB))
		require.Equal(t, server.DedicatedCPU, desiredServer.ServerPlanCommitment.IsDedicatedCPU())
	})
}
