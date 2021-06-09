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
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/log"
	"google.golang.org/grpc"
)

type Server interface {
	Name() string
	Version() string
	GetLogger() *log.Logger
	SetLogger(logger *log.Logger)
}

type Handler interface {
	Handle(*handler.HandleRequest, ResponseSender) error
}

type PreHandler interface {
	PreHandle(*handler.PreHandleRequest, ResponseSender) error
}

type PostHandler interface {
	PostHandle(*handler.PostHandleRequest, ResponseSender) error
}

type ResponseSender interface {
	Send(*handler.HandleResponse) error
}

type FlagCustomizer interface {
	CustomizeFlags(fs *flag.FlagSet)
}

func Serve(server Server) {
	handlerName := HandlerFullName(server)
	logger := server.GetLogger().With("from", handlerName)

	validateHandlerInterfaces(server, logger)

	fs := flag.CommandLine
	var address string
	fs.StringVar(&address, "address", fmt.Sprintf("%s.sock", handlerName), "URL of gRPC endpoint of the handler")

	var showHelp, showVersion bool
	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Show version")

	// 各Handlerでのカスタマイズ
	if fc, ok := server.(FlagCustomizer); ok {
		fc.CustomizeFlags(fs)
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal("fatal", err)
	}

	// TODO add flag validation

	switch {
	case showHelp:
		showUsage(handlerName, fs)
		return
	case showVersion:
		fmt.Println(server.Version())
		return
	default:
		errCh := make(chan error)
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// TODO 簡易的な実装、後ほど整理&切り出し
		filename := strings.Replace(address, "unix:", "", -1)
		lis, err := net.Listen("unix", filename)
		if err != nil {
			logger.Fatal("fatal", err)
		}

		grpcServer := grpc.NewServer()
		srv := &HandleService{
			Handler: server,
			// TODO ロガーの設定
			logger: log.NewLogger(&log.LoggerOption{
				Writer:    nil,
				JSON:      false,
				TimeStamp: true,
				Caller:    false,
				Level:     log.LevelInfo,
			}),
		}
		handler.RegisterHandleServiceServer(grpcServer, srv)

		defer func() {
			grpcServer.GracefulStop()
			lis.Close()
			if _, err := os.Stat(filename); err == nil {
				if err := os.RemoveAll(filename); err != nil {
					logger.Fatal("fatal", err) // nolint
				}
			}
		}()

		go func() {
			if err := logger.Info("message", "started", "address", lis.Addr().String()); err != nil {
				errCh <- err
			}
			if err := grpcServer.Serve(lis); err != nil {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			logger.Fatal("fatal", err)
		case <-ctx.Done():
			logger.Info("message", "shutting down", "error", ctx.Err()) // nolint
		}
	}
}

func validateHandlerInterfaces(server Server, logger *log.Logger) {
	if _, ok := server.(PreHandler); ok {
		return
	}
	if _, ok := server.(Handler); ok {
		return
	}
	if _, ok := server.(PostHandler); ok {
		return
	}
	logger.Fatal("fatal", "At least one of the following must be implemented: PreHandler or Handler or PostHandler") // nolint
}

func showUsage(name string, fs *flag.FlagSet) {
	fmt.Printf("usage: %s [flags]\n", name)
	fs.Usage()
}

func HandlerFullName(server Server) string {
	return fmt.Sprintf("autoscaler-handlers-%s", server.Name())
}

var _ handler.HandleServiceServer = (*HandleService)(nil)

type HandleService struct {
	handler.UnimplementedHandleServiceServer
	Handler Server
	logger  *log.Logger
}

func (h *HandleService) PreHandle(req *handler.PreHandleRequest, server handler.HandleService_PreHandleServer) error {
	if impl, ok := h.Handler.(PreHandler); ok {
		if err := h.logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := h.logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PreHandle(req, server)
	}

	if err := h.logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return h.logger.Debug("request", req.String())
}

func (h *HandleService) Handle(req *handler.HandleRequest, server handler.HandleService_HandleServer) error {
	if impl, ok := h.Handler.(Handler); ok {
		if err := h.logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := h.logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.Handle(req, server)
	}

	if err := h.logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return h.logger.Debug("request", req.String())
}

func (h *HandleService) PostHandle(req *handler.PostHandleRequest, server handler.HandleService_PostHandleServer) error {
	if impl, ok := h.Handler.(PostHandler); ok {
		if err := h.logger.Info("status", handler.HandleResponse_RECEIVED); err != nil {
			return err
		}
		if err := h.logger.Debug("request", req.String()); err != nil {
			return err
		}
		return impl.PostHandle(req, server)
	}

	if err := h.logger.Info("status", handler.HandleResponse_IGNORED); err != nil {
		return err
	}
	return h.logger.Debug("request", req.String())
}
