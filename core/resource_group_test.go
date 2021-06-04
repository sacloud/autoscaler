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
	"fmt"
	"reflect"
	"testing"

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
	}

	type fields struct {
		HandlerConfigs []*ResourceHandlerConfig
		Resources      Resources
		Name           string
	}
	type args struct {
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
			name: "returns if HandlerConfigs is empty",
			fields: fields{
				HandlerConfigs: nil,
				Name:           "empty",
			},
			args: args{
				allHandlers: allHandlers,
			},
			want:    allHandlers,
			wantErr: false,
		},
		{
			name: "returns error if invalid HandlerConfigs is specified",
			fields: fields{
				HandlerConfigs: []*ResourceHandlerConfig{
					{
						Name: "not-exists",
					},
				},
				Name: "not-exists",
			},
			args: args{
				allHandlers: allHandlers,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns with filtering by HandlerConfigs",
			fields: fields{
				HandlerConfigs: []*ResourceHandlerConfig{
					{
						Name: "dummy1",
					},
				},
				Name: "filter",
			},
			args: args{
				allHandlers: allHandlers,
			},
			want: Handlers{
				{
					Type: "dummy",
					Name: "dummy1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := &ResourceGroup{
				HandlerConfigs: tt.fields.HandlerConfigs,
				Resources:      tt.fields.Resources,
				Name:           tt.fields.Name,
			}
			got, err := rg.handlers(tt.args.allHandlers)
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
			HandlerConfigs: nil,
			Resources: Resources{
				&stubResource{
					ResourceBase: &ResourceBase{},
					computeFunc: func(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
						called++
						return []Computed{&stubComputed{
							instruction: handler.ResourceInstructions_NOOP,
							current:     &handler.Resource{},
							desired:     &handler.Resource{},
						}}, nil
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

	t.Run("calls PreHandle/Handle/PostHandle for each Computed", func(t *testing.T) {
		var history []string
		rg := &ResourceGroup{
			HandlerConfigs: nil,
			Resources: Resources{
				&stubResource{
					ResourceBase: &ResourceBase{},
					computeFunc: func(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
						return []Computed{
							&stubComputed{
								instruction: handler.ResourceInstructions_NOOP,
								current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "1"}}},
								desired:     &handler.Resource{},
							},
							&stubComputed{
								instruction: handler.ResourceInstructions_NOOP,
								current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "2"}}},
								desired:     &handler.Resource{},
							},
						}, nil
					},
				},
				&stubResource{
					ResourceBase: &ResourceBase{},
					computeFunc: func(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
						return []Computed{
							&stubComputed{
								instruction: handler.ResourceInstructions_NOOP,
								current:     &handler.Resource{Resource: &handler.Resource_Server{Server: &handler.Server{Id: "3"}}},
								desired:     &handler.Resource{},
							},
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
						history = append(history, fmt.Sprintf("Handler1->PreHandle:%s", request.Current.GetServer().Id))
						return nil
					},
					HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
						history = append(history, fmt.Sprintf("Handler1->Handle:%s", request.Current.GetServer().Id))
						return nil
					},
					PostHandleFunc: func(request *handler.PostHandleRequest, sender handlers.ResponseSender) error {
						history = append(history, fmt.Sprintf("Handler1->PostHandle:%s", request.Current.GetServer().Id))
						return nil
					},
				},
			},
			{
				Type: "stub",
				Name: "stub",
				BuiltinHandler: &stub.Handler{
					PreHandleFunc: func(request *handler.PreHandleRequest, sender handlers.ResponseSender) error {
						history = append(history, fmt.Sprintf("Handler2->PreHandle:%s", request.Current.GetServer().Id))
						return nil
					},
					HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
						history = append(history, fmt.Sprintf("Handler2->Handle:%s", request.Current.GetServer().Id))
						return nil
					},
					PostHandleFunc: func(request *handler.PostHandleRequest, sender handlers.ResponseSender) error {
						history = append(history, fmt.Sprintf("Handler2->PostHandle:%s", request.Current.GetServer().Id))
						return nil
					},
				},
			},
		})

		expected := []string{
			// PreHandle/PostHandleはCompute()が返した[]Computedごと
			"Handler1->PreHandle:1", "Handler2->PreHandle:1",
			"Handler1->PreHandle:2", "Handler2->PreHandle:2",

			"Handler1->Handle:1", "Handler2->Handle:1",
			"Handler1->Handle:2", "Handler2->Handle:2",

			// PostHandleはその後
			"Handler1->PostHandle:1", "Handler2->PostHandle:1",
			"Handler1->PostHandle:2", "Handler2->PostHandle:2",

			// 別Resourceが返した[]Computedはその後
			"Handler1->PreHandle:3", "Handler2->PreHandle:3",
			"Handler1->Handle:3", "Handler2->Handle:3",
			"Handler1->PostHandle:3", "Handler2->PostHandle:3",
		}

		require.Equal(t, expected, history)
	})
}
