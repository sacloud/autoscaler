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
	"reflect"
	"testing"

	"github.com/sacloud/autoscaler/defaults"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/builtins"
	"github.com/sacloud/autoscaler/handlers/stub"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestResourceGroup_handlers(t *testing.T) {
	allHandlers := Handlers{
		{
			Type: "dummy",
			Name: "dummy1",
		},
		{
			Type: "dummy",
			Name: "dummy2",
		},
		{
			Type:     "dummy",
			Name:     "dummy3",
			Disabled: true,
		},
	}

	type fields struct {
		Actions   Actions
		Resources Resources
		Name      string
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
					Type: "dummy",
					Name: "dummy1",
				},
				{
					Type: "dummy",
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
					Type:     "dummy",
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
					Type: "dummy",
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
					Type: "dummy",
					Name: "dummy2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := &ResourceGroup{
				Actions:   tt.fields.Actions,
				Resources: tt.fields.Resources,
				Name:      tt.fields.Name,
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

func TestResourceGroup_handleAll(t *testing.T) {
	t.Run("calls Compute() func twice", func(t *testing.T) {
		called := 0
		rg := &ResourceGroup{
			Resources: Resources{
				&stubResource{
					ResourceBase: &ResourceBase{},
					computeFunc: func(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
						called++
						return &stubComputed{
							instruction: handler.ResourceInstructions_NOOP,
							current:     &handler.Resource{},
							desired:     &handler.Resource{},
						}, nil
					},
				},
			},
			Name: "test",
		}

		rg.handleAll(testContext(), testAPIClient, Handlers{ // nolint
			{
				Type: "stub",
				Name: "stub",
				BuiltinHandler: &builtins.Handler{
					Builtin: &stub.Handler{
						PreHandleFunc: func(request *handler.PreHandleRequest, sender handlers.ResponseSender) error {
							return nil
						},
						HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
							return nil
						},
						PostHandleFunc: func(request *handler.PostHandleRequest, sender handlers.ResponseSender) error {
							return nil
						},
					},
				},
			},
		})

		// handleAll中にCompute()が2回(初回+リフレッシュ)呼ばれているか?
		require.Equal(t, 2, called)
	})

	t.Run("compute current/desired state with parent", func(t *testing.T) {
		var history []string
		rg := &ResourceGroup{
			Resources: Resources{
				&stubResource{
					ResourceBase: &ResourceBase{
						Children: Resources{
							&stubResource{
								ResourceBase: &ResourceBase{
									Children: Resources{
										&stubResource{
											ResourceBase: &ResourceBase{},
											computeFunc: func(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
												history = append(history, "child2")
												return &stubComputed{
													instruction: handler.ResourceInstructions_NOOP,
													current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "3"}}},
													desired:     &handler.Resource{},
												}, nil
											},
										},
									},
								},
								computeFunc: func(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
									history = append(history, "child1")
									return &stubComputed{
										instruction: handler.ResourceInstructions_NOOP,
										current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "2"}}},
										desired:     &handler.Resource{},
									}, nil
								},
							},
						},
					},
					computeFunc: func(ctx *Context, apiClient sacloud.APICaller) (Computed, error) {
						history = append(history, "parent")
						return &stubComputed{
							instruction: handler.ResourceInstructions_NOOP,
							current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "1"}}},
							desired:     &handler.Resource{},
						}, nil
					},
				},
			},
			Name: "test",
		}

		rg.handleAll(testContext(), testAPIClient, Handlers{ // nolint
			{
				Type: "stub",
				Name: "stub",
				BuiltinHandler: &stub.Handler{
					PreHandleFunc: func(request *handler.PreHandleRequest, sender handlers.ResponseSender) error {
						return nil
					},
					HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
						return nil
					},
					PostHandleFunc: func(request *handler.PostHandleRequest, sender handlers.ResponseSender) error {
						return nil
					},
				},
			},
		})

		expected := []string{
			"parent", // 親のCompute
			"child1", // 子1のCompute
			"child2", // 子2のCompute
			"child2", // 子2のRefresh
			"child1", // 子のRefresh
			"parent", // 親のRefresh
		}
		require.Equal(t, expected, history)
	})
}
