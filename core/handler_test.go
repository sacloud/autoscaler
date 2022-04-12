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

	"github.com/sacloud/autoscaler/handler"
	"github.com/stretchr/testify/require"
)

func Test_handleHandlerResponseStatus_eachStatus(t *testing.T) {
	tests := map[handler.HandleResponse_Status]bool{
		handler.HandleResponse_UNKNOWN:  false,
		handler.HandleResponse_RECEIVED: false,
		handler.HandleResponse_ACCEPTED: false,
		handler.HandleResponse_RUNNING:  true,
		handler.HandleResponse_DONE:     true,
		handler.HandleResponse_IGNORED:  false,
	}
	for status, want := range tests {
		ctx := &HandlingContext{
			RequestContext: &RequestContext{
				handled: false,
			},
		}
		handleHandlerResponseStatus(ctx, status)
		require.Equal(t, want, ctx.handled)
	}
}

func Test_handleHandlerResponseStatus_setTrueOnce(t *testing.T) {
	ctx := &HandlingContext{
		RequestContext: &RequestContext{
			handled: false,
		},
	}
	handleHandlerResponseStatus(ctx, handler.HandleResponse_UNKNOWN)
	require.False(t, ctx.handled)

	handleHandlerResponseStatus(ctx, handler.HandleResponse_RECEIVED)
	require.False(t, ctx.handled)

	handleHandlerResponseStatus(ctx, handler.HandleResponse_RUNNING)
	require.True(t, ctx.handled)

	// 一度TrueになったらFalseになることはない
	handleHandlerResponseStatus(ctx, handler.HandleResponse_RECEIVED)
	require.True(t, ctx.handled)
}
