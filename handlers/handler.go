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

package handlers

import (
	"context"
	"fmt"
)

// Serve 指定のハンドラでgRPCサーバをスタート/リッスンする
func Serve(parentCtx context.Context, server HandlerMeta) error {
	server.SetLogger(server.GetLogger().With("from", handlerFullName(server)))

	validateHandlerInterfaces(server)

	service := &handleService{
		Handler: server,
	}
	return service.listenAndServe(parentCtx)
}

func handlerFullName(server HandlerMeta) string {
	return fmt.Sprintf("autoscaler-handlers-%s", server.Name())
}

func validateHandlerInterfaces(server HandlerMeta) {
	if _, ok := server.(Listener); !ok {
		server.GetLogger().Fatal("fatal", "Handler must be implemented Listener interface") // nolint
	}

	if _, ok := server.(PreHandler); ok {
		return
	}
	if _, ok := server.(Handler); ok {
		return
	}
	if _, ok := server.(PostHandler); ok {
		return
	}
	server.GetLogger().Fatal("fatal", "At least one of the following must be implemented: PreHandler or Handler or PostHandler") // nolint
}

