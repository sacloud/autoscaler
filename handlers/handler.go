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

package handlers

import (
	"context"
	"fmt"
	"net"

	"github.com/sacloud/autoscaler/metrics"
)

const (
	HandlerTypePreHandle  = "pre-handle"
	HandlerTypeHandle     = "handle"
	HandlerTypePostHandle = "post-handle"
)

// Serve 指定のハンドラでgRPCサーバをスタート/リッスンする
func Serve(ctx context.Context, handler CustomHandler) error {
	handler.SetLogger(handler.GetLogger().With("from", handlerFullName(handler)))
	validateHandlerInterfaces(handler)

	conf, err := LoadConfigFromPath(handler.ConfigPath())
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		if err := startHandlerService(ctx, handler, conf); err != nil {
			errCh <- err
		}
	}()
	go func() {
		if err := startExporter(ctx, handler, conf); err != nil {
			errCh <- err
		}
	}()

	logger := handler.GetLogger()
	for {
		select {
		case err := <-errCh:
			logger.Error("error", err) // nolint
		case <-ctx.Done():
			logger.Info("message", "shutting down", "error", ctx.Err()) // nolint
			return ctx.Err()
		}
	}
}

func startHandlerService(ctx context.Context, handler CustomHandler, conf *Config) error {
	service := &handleService{
		Handler: handler,
		conf:    conf,
	}
	return service.listenAndServe(ctx)
}

func startExporter(ctx context.Context, handler CustomHandler, conf *Config) error {
	if conf != nil && conf.ExporterConfig != nil && conf.ExporterConfig.Enabled {
		listener, err := net.Listen("tcp", conf.ExporterConfig.Address)
		if err != nil {
			return err
		}

		server := metrics.NewServer(conf.ExporterConfig.Address, conf.ExporterConfig.TLSConfig, handler.GetLogger())
		defer func() {
			server.Shutdown(ctx) // nolint
			listener.Close()
		}()

		return server.Serve(listener)
	}
	return nil
}

func handlerFullName(server HandlerMeta) string {
	return fmt.Sprintf("autoscaler-handlers-%s", server.Name())
}

func validateHandlerInterfaces(server HandlerMeta) {
	if _, ok := server.(Listener); !ok {
		server.GetLogger().Fatal("fatal", "Handler must be implemented Listener interface")
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
	server.GetLogger().Fatal("fatal", "At least one of the following must be implemented: PreHandler or Handler or PostHandler")
}
