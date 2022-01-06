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
	"testing"

	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

func TestGraph_Tree(t *testing.T) {
	ctx := testContext()
	client := test.APIClient

	type fields struct {
		resources ResourceDefinitions
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				resources: ResourceDefinitions{
					&stubResourceDef{
						ResourceDefBase: &ResourceDefBase{
							TypeName: "Stub",
							children: ResourceDefinitions{
								&stubResourceDef{
									ResourceDefBase: &ResourceDefBase{
										TypeName: "Stub",
									},
									computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
										return Resources{
											&stubResource{
												ResourceBase: &ResourceBase{
													resourceType: ResourceTypeServer,
												},
												computeFunc: func(ctx *RequestContext, refresh bool) (Computed, error) {
													return nil, nil
												},
											},
										}, nil
									},
								},
							},
						},
						computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
							return Resources{
								&stubResource{
									ResourceBase: &ResourceBase{
										resourceType: ResourceTypeDNS,
									},
									computeFunc: func(ctx *RequestContext, refresh bool) (Computed, error) {
										return nil, nil
									},
								},
							}, nil
						},
					},
				},
			},
			want: `
Sacloud AutoScaler
└─ stub
   └─ stub
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Graph{
				resources: tt.fields.resources,
			}
			got, err := g.Tree(ctx, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ("\n" + got) != tt.want {
				t.Errorf("Tree() got = \n%v, want %v", got, tt.want)
			}
		})
	}
}
