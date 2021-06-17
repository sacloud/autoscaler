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
	"reflect"
	"testing"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestResourceDefGroup_handlers(t *testing.T) {
	allHandlers := Handlers{
		{
			Name: "dummy1",
		},
		{
			Name: "dummy2",
		},
		{
			Name:     "dummy3",
			Disabled: true,
		},
	}

	type fields struct {
		Actions Actions
		Name    string
	}
	type args struct {
		actionName  string
		allHandlers Handlers
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Handlers
		wantErr bool
	}{
		{
			name: "returns all enabled handlers if Actions is empty",
			fields: fields{
				Name:    "empty",
				Actions: Actions{},
			},
			args: args{
				actionName:  defaults.ActionName,
				allHandlers: allHandlers,
			},
			want: Handlers{
				{
					Name: "dummy1",
				},
				{
					Name: "dummy2",
				},
			},
			wantErr: false,
		},
		{
			name: "returns error if invalid ActionName is specified",
			fields: fields{
				Actions: Actions{
					"foobar":   []string{"dummy1", "dummy2"},
					"disabled": []string{"dummy1", "dummy2", "dummy3"},
				},
				Name: "not-exists",
			},
			args: args{
				allHandlers: allHandlers,
				actionName:  "not-exists",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns error with invalid definition of Actions",
			fields: fields{
				Actions: Actions{
					"foobar": []string{},
				},
				Name: "foobar",
			},
			args: args{
				allHandlers: allHandlers,
				actionName:  "foobar",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns handlers even at Disabled:true",
			fields: fields{
				Actions: Actions{
					"filter": []string{"dummy3"},
				},
				Name: "filter",
			},
			args: args{
				allHandlers: allHandlers,
				actionName:  "filter",
			},
			want: Handlers{
				{
					Name:     "dummy3",
					Disabled: true,
				},
			},
			wantErr: false,
		},
		{
			name: "returns first handlers if action name is empty",
			fields: fields{
				Actions: Actions{
					"action1": []string{"dummy2"},
				},
				Name: "filter",
			},
			args: args{
				allHandlers: allHandlers,
				actionName:  "",
			},
			want: Handlers{
				{
					Name: "dummy2",
				},
			},
			wantErr: false,
		},
		{
			name: "returns first handlers if action name is default value",
			fields: fields{
				Actions: Actions{
					"action1": []string{"dummy2"},
				},
				Name: "filter",
			},
			args: args{
				allHandlers: allHandlers,
				actionName:  defaults.ActionName,
			},
			want: Handlers{
				{
					Name: "dummy2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := &ResourceDefGroup{
				Actions: tt.fields.Actions,
				name:    tt.fields.Name,
			}
			got, err := rg.handlers(tt.args.actionName, tt.args.allHandlers)
			if (err != nil) != tt.wantErr {
				t.Errorf("handlers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handlers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceDefGroup_ResourceGroup(t *testing.T) {
	type fields struct {
		ResourceDefs ResourceDefinitions
	}
	type args struct {
		ctx       *RequestContext
		apiClient sacloud.APICaller
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ResourceGroup2
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				// DNS
				//  |-- ELB1
				//  |    |-- Server1
				//  |    |-- Server2
				//  |-- ELB2
				//       |-- Server3
				//       |-- Server4
				// GSLB
				//  |------- Server5
				//  |------- Server6
				ResourceDefs: ResourceDefinitions{
					&stubResourceDef{
						computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
							return Resources2{
								&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeDNS, nil)},
							}, nil
						},
						ResourceDefBase: &ResourceDefBase{
							TypeName: ResourceTypeDNS.String(),
							children: ResourceDefinitions{
								&stubResourceDef{
									computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
										return Resources2{
											&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeEnhancedLoadBalancer, nil)},
										}, nil
									},
									ResourceDefBase: &ResourceDefBase{
										TypeName: ResourceTypeEnhancedLoadBalancer.String(),
										children: ResourceDefinitions{
											&stubResourceDef{
												computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
													return Resources2{
														&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
													}, nil
												},
												ResourceDefBase: &ResourceDefBase{
													TypeName: ResourceTypeServer.String(),
												},
											},
											&stubResourceDef{
												computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
													return Resources2{
														&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
													}, nil
												},
												ResourceDefBase: &ResourceDefBase{
													TypeName: ResourceTypeServer.String(),
												},
											},
										},
									},
								},
								&stubResourceDef{
									computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
										return Resources2{
											&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeEnhancedLoadBalancer, nil)},
										}, nil
									},
									ResourceDefBase: &ResourceDefBase{
										TypeName: ResourceTypeEnhancedLoadBalancer.String(),
										children: ResourceDefinitions{
											&stubResourceDef{
												computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
													return Resources2{
														&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
													}, nil
												},
												ResourceDefBase: &ResourceDefBase{
													TypeName: ResourceTypeServer.String(),
												},
											},
											&stubResourceDef{
												computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
													return Resources2{
														&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
													}, nil
												},
												ResourceDefBase: &ResourceDefBase{
													TypeName: ResourceTypeServer.String(),
												},
											},
										},
									},
								},
							},
						},
					},
					&stubResourceDef{
						computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
							return Resources2{
								&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeGSLB, nil)},
							}, nil
						},
						ResourceDefBase: &ResourceDefBase{
							TypeName: ResourceTypeGSLB.String(),
							children: ResourceDefinitions{
								&stubResourceDef{
									computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
										return Resources2{
											&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
										}, nil
									},
									ResourceDefBase: &ResourceDefBase{
										TypeName: ResourceTypeServer.String(),
									},
								},
								&stubResourceDef{
									computeFunc: func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
										return Resources2{
											&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
										}, nil
									},
									ResourceDefBase: &ResourceDefBase{
										TypeName: ResourceTypeServer.String(),
									},
								},
							},
						},
					},
				},
			},
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType:       requestTypeUp,
						source:            "default",
						action:            "default",
						resourceGroupName: "default",
					},
					job: &JobStatus{
						requestType: requestTypeUp,
						id:          "default-default-default",
					},
					logger: test.Logger,
				},
				apiClient: test.APIClient,
			},
			want: &ResourceGroup2{
				Resources: Resources2{
					&stubResource2{
						ResourceBase2: NewResourceBase2(
							ResourceTypeDNS,
							nil,
							&stubResource2{
								ResourceBase2: NewResourceBase2(
									ResourceTypeEnhancedLoadBalancer, nil,
									&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
									&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
								),
							},
							&stubResource2{
								ResourceBase2: NewResourceBase2(
									ResourceTypeEnhancedLoadBalancer, nil,
									&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
									&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
								),
							},
						),
					},
					&stubResource2{
						ResourceBase2: NewResourceBase2(
							ResourceTypeGSLB,
							nil,
							&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
							&stubResource2{ResourceBase2: NewResourceBase2(ResourceTypeServer, nil)},
						),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := &ResourceDefGroup{
				ResourceDefs: tt.fields.ResourceDefs,
			}
			got, err := rg.ResourceGroup(tt.args.ctx, tt.args.apiClient)
			require.Equal(t, tt.wantErr, err != nil)
			require.EqualValues(t, tt.want, got)
		})
	}
}
