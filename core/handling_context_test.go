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
	"testing"

	"github.com/sacloud/autoscaler/handler"
)

func TestHandlingContext_ComputeResult(t *testing.T) {
	type fields struct {
		Context         *RequestContext
		currentComputed Computed
	}
	type args struct {
		computed Computed
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   handler.PostHandleRequest_ResourceHandleResults
	}{
		{
			name: "created",
			fields: fields{
				currentComputed: &stubComputed{
					id:          "",
					instruction: handler.ResourceInstructions_CREATE,
				},
			},
			args: args{
				computed: &stubComputed{
					id:          "1",
					instruction: handler.ResourceInstructions_NOOP,
				},
			},
			want: handler.PostHandleRequest_CREATED,
		},
		{
			name: "updated",
			fields: fields{
				currentComputed: &stubComputed{
					id:          "1",
					instruction: handler.ResourceInstructions_UPDATE,
				},
			},
			args: args{
				computed: &stubComputed{
					id:          "1",
					instruction: handler.ResourceInstructions_NOOP,
				},
			},
			want: handler.PostHandleRequest_UPDATED,
		},
		{
			name: "deleted",
			fields: fields{
				currentComputed: &stubComputed{
					id:          "1",
					instruction: handler.ResourceInstructions_DELETE,
				},
			},
			args: args{
				computed: &stubComputed{
					id:          "",
					instruction: handler.ResourceInstructions_NOOP,
				},
			},
			want: handler.PostHandleRequest_DELETED,
		},
		{
			name: "plan changed",
			fields: fields{
				currentComputed: &stubComputed{
					id:          "1",
					instruction: handler.ResourceInstructions_UPDATE,
				},
			},
			args: args{
				computed: &stubComputed{
					id:          "2",
					instruction: handler.ResourceInstructions_NOOP,
				},
			},
			want: handler.PostHandleRequest_UPDATED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &HandlingContext{
				RequestContext: tt.fields.Context,
				cachedComputed: tt.fields.currentComputed,
			}
			if got := c.ComputeResult(tt.args.computed); got != tt.want {
				t.Errorf("ComputeResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
