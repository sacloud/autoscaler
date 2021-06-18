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

	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestResourceDefServer_Validate(t *testing.T) {
	defer initTestServer(t)()

	t.Run("returns error if selector is empty", func(t *testing.T) {
		empty := &Server{
			ResourceBase: &ResourceBase{TypeName: "Server"},
		}
		errs := empty.Validate(context.Background(), test.APIClient)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "resource=Server: selector: required")
	})

	t.Run("returns error if selector.Zone is empty", func(t *testing.T) {
		empty := &Server{
			ResourceBase: &ResourceBase{
				TypeName:       "Server",
				TargetSelector: &ResourceSelector{},
			},
		}
		errs := empty.Validate(context.Background(), test.APIClient)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "resource=Server: selector.Zone: required")
	})

	t.Run("returns error if servers were not found", func(t *testing.T) {
		empty := &Server{
			ResourceBase: &ResourceBase{
				TypeName: "Server",
				TargetSelector: &ResourceSelector{
					Zone:  "is1a",
					Names: []string{"server-not-found"},
				},
			},
		}
		errs := empty.Validate(context.Background(), test.APIClient)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "resource=Server: resource not found with selector: ID: , Names: [server-not-found], Zone: is1a")
	})
}

func TestResourceDefServer_Compute(t *testing.T) {
	defer initTestServer(t)()

	type fields struct {
		ResourceDefBase *ResourceDefBase
	}
	type args struct {
		ctx       *RequestContext
		apiClient sacloud.APICaller
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Resources2
		wantErr bool
	}{
		{
			name: "simple",
			fields: fields{
				ResourceDefBase: &ResourceDefBase{
					TypeName: ResourceTypeServer.String(),
					TargetSelector: &ResourceSelector{
						Names: []string{"test-server"},
						Zone:  test.Zone,
					},
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
