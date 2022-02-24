// Copyright 2021-2022 The sacloud/autoscaler Authors
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
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestResourceDefServer_Validate(t *testing.T) {
	validate.InitValidatorAlias(sacloud.SakuraCloudZones)
	_, cleanup := test.AddTestServer(t, "test-server")
	defer cleanup()

	t.Run("returns error if servers were not found", func(t *testing.T) {
		empty := &ResourceDefServer{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "Server",
			},
			Selector: &MultiZoneSelector{
				ResourceSelector: &ResourceSelector{
					Names: []string{"server-not-found"},
				},
				Zones: []string{"is1a"},
			},
		}
		errs := empty.Validate(context.Background(), test.APIClient)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "resource=Server resource not found with selector: ID: , Names: [server-not-found], Tags: [], Zones: [is1a]")
	})
}

func TestResourceDefServer_Compute(t *testing.T) {
	_, cleanup := test.AddTestServer(t, "test-server")
	defer cleanup()

	type fields struct {
		ResourceDefBase *ResourceDefBase
		Selector        *MultiZoneSelector
	}
	type args struct {
		ctx       *RequestContext
		apiClient sacloud.APICaller
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Resources
		wantErr bool
	}{
		{
			name: "simple",
			fields: fields{
				ResourceDefBase: &ResourceDefBase{
					TypeName: ResourceTypeServer.String(),
				},
				Selector: &MultiZoneSelector{
					ResourceSelector: &ResourceSelector{
						Names: []string{"test-server"},
					},
					Zones: []string{test.Zone},
				},
			},
			args: args{
				ctx: &RequestContext{
					ctx:    context.Background(),
					logger: test.Logger,
				},
				apiClient: test.APIClient,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ResourceDefServer{
				ResourceDefBase: tt.fields.ResourceDefBase,
				Selector:        tt.fields.Selector,
			}
			got, err := s.Compute(tt.args.ctx, tt.args.apiClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Len(t, got, 1)
		})
	}
}

func TestServer_ComputedWithResource(t *testing.T) {
	_, cleanup := test.AddTestServer(t, "test-server")
	defer cleanup()

	ctx := testContext()

	t.Run("returns error if selector has invalid value", func(t *testing.T) {
		notFound := &ResourceDefServer{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "Server",
				DefName:  "default",
			},
			Selector: &MultiZoneSelector{
				ResourceSelector: &ResourceSelector{
					ID: 123456789012,
				},
				Zones: []string{test.Zone},
			},
		}

		_, err := notFound.Compute(ctx, test.APIClient)
		require.Error(t, err)
	})

	t.Run("returns UPDATE instruction if selector has valid value", func(t *testing.T) {
		running := &ResourceDefServer{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "Server",
				DefName:  "default",
			},
			Selector: &MultiZoneSelector{
				ResourceSelector: &ResourceSelector{
					Names: []string{"test-server"},
				},
				Zones: []string{test.Zone},
			},
			Plans: []*ServerPlan{
				{Core: 1, Memory: 1},
				{Core: 2, Memory: 4},
				{Core: 4, Memory: 8},
			},
		}

		resources, err := running.Compute(ctx, test.APIClient)
		require.NoError(t, err)
		require.Len(t, resources, 1)
		resource := resources[0]

		computed, err := resource.Compute(ctx, false)
		require.NoError(t, err)
		require.NotNil(t, computed)

		require.Equal(t, handler.ResourceInstructions_UPDATE, computed.Instruction())

		current := computed.Current()
		require.NotNil(t, current)

		desired := computed.Desired()
		require.NotNil(t, desired)
	})

	t.Run("returns desired state that can convert to the request parameter", func(t *testing.T) {
		ctx := testContext()
		server := &ResourceDefServer{
			ResourceDefBase: &ResourceDefBase{
				TypeName: "Server",
				DefName:  "default",
			},
			Selector: &MultiZoneSelector{
				ResourceSelector: &ResourceSelector{
					Names: []string{"test-server"},
				},
				Zones: []string{test.Zone},
			},
			Plans: []*ServerPlan{
				{Core: 1, Memory: 1},
				{Core: 2, Memory: 4},
				{Core: 4, Memory: 8},
			},
		}
		resources, err := server.Compute(ctx, test.APIClient)
		require.NoError(t, err)
		computed, err := resources[0].Compute(ctx, false)
		require.NoError(t, err)

		handlerReq := computed.Desired()
		require.NotNil(t, handlerReq)

		desiredServer := handlerReq.GetServer()
		require.NotNil(t, desiredServer)

		// Server.Plansで指定した次のプランが返されるはず
		require.Equal(t, uint32(4), desiredServer.Core)
		require.Equal(t, uint32(8), desiredServer.Memory)
		require.Equal(t, server.DedicatedCPU, desiredServer.DedicatedCpu)
	})
}
