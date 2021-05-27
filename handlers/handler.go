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
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sacloud/autoscaler/handler"
	"google.golang.org/grpc"
)

type Server interface {
	Name() string
	Version() string
	Handle(*handler.HandleRequest, ResponseSender) error
}

type ResponseSender interface {
	Send(*handler.HandleResponse) error
}

type FlagCustomizer interface {
	CustomizeFlags(fs *flag.FlagSet)
}

func Serve(server Server) {
	handlerName := fmt.Sprintf("autoscaler-handlers-%s", server.Name())

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
		log.Fatal(err)
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
			log.Fatal(err)
		}

		grpcServer := grpc.NewServer()
		srv := &handleService{
			handler: server,
		}
		handler.RegisterHandleServiceServer(grpcServer, srv)

		defer func() {
			grpcServer.GracefulStop()
			lis.Close()
			if _, err := os.Stat(filename); err == nil {
				if err := os.RemoveAll(filename); err != nil {
					log.Fatal(err)
				}
			}
		}()

		go func() {
			log.Printf("%s started with: %s\n", handlerName, lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			log.Fatalln("Fatal error: ", err)
		case <-ctx.Done():
			log.Println("shutting down with:", ctx.Err())
		}
	}
}

func showUsage(name string, fs *flag.FlagSet) {
	fmt.Printf("usage: %s [flags]\n", name)
	fs.Usage()
}

var _ handler.HandleServiceServer = (*handleService)(nil)

type handleService struct {
	handler.UnimplementedHandleServiceServer
	handler Server
}

func (h *handleService) Handle(req *handler.HandleRequest, server handler.HandleService_HandleServer) error {
	return h.handler.Handle(req, server)
}
