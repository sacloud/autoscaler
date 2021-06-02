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

package inputs

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
)

type Input interface {
	Name() string
	Version() string
	ShouldAccept(req *http.Request) (bool, error) // true,nilを返した場合のみCoreへのリクエストを行う
}

type FlagCustomizer interface {
	CustomizeFlags(fs *flag.FlagSet)
}

func showUsage(name string, fs *flag.FlagSet) {
	fmt.Printf("usage: %s [flags]\n", name)
	fs.Usage()
}

func FullName(input Input) string {
	return fmt.Sprintf("autoscaler-inputs-%s", input.Name())
}

func Serve(input Input) {
	name := FullName(input)

	fs := flag.CommandLine
	var dest, address string
	flag.StringVar(&dest, "dest", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	flag.StringVar(&address, "addr", ":3001", "the TCP address for the server to listen on")

	var showHelp, showVersion, debug bool
	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Show version")
	fs.BoolVar(&debug, "debug", false, "Show debug logs")

	// 各Handlerでのカスタマイズ
	if fc, ok := input.(FlagCustomizer); ok {
		fc.CustomizeFlags(fs)
	}

	// TODO add flag validation

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	switch {
	case showHelp:
		showUsage(name, fs)
		return
	case showVersion:
		fmt.Println(input.Version())
		return
	default:
		server := &server{
			coreAddress:   dest,
			listenAddress: address,
			input:         input,
			debug:         debug,
		}
		if err := server.listenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
}

type server struct {
	coreAddress   string
	listenAddress string
	input         Input
	debug         bool
}

func (s *server) listenAndServe() error {
	serveMux := http.DefaultServeMux

	serveMux.HandleFunc("/up", func(w http.ResponseWriter, req *http.Request) {
		s.handle("up", w, req)
	})
	serveMux.HandleFunc("/down", func(w http.ResponseWriter, req *http.Request) {
		s.handle("down", w, req)
	})
	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) // nolint
	})

	log.Printf("%s: started on %s\n", FullName(s.input), s.listenAddress)
	return http.ListenAndServe(s.listenAddress, serveMux)
}

func (s *server) handle(requestType string, w http.ResponseWriter, req *http.Request) {
	scalingReq, err := s.parseRequest(requestType, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if scalingReq == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.send(scalingReq); err != nil {
		log.Println("[ERROR]: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) // nolint
}

func (s *server) parseRequest(requestType string, req *http.Request) (*scalingRequest, error) {
	log.Println("webhook received")
	if s.debug {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		log.Println(string(dump))
	}

	shouldAccept, err := s.input.ShouldAccept(req)
	if err != nil {
		return nil, err
	}
	if !shouldAccept {
		log.Println("webhook ignored")
		return nil, nil
	}

	queryStrings := req.URL.Query()
	source := queryStrings.Get("source")
	if source == "" {
		source = defaults.SourceName
	}
	action := queryStrings.Get("action")
	if action == "" {
		action = defaults.ActionName
	}
	groupName := queryStrings.Get("resource_group_name")
	if groupName == "" {
		groupName = defaults.ResourceGroupName
	}
	return &scalingRequest{
		source:      source,
		action:      action,
		groupName:   groupName,
		requestType: requestType,
	}, nil
}

func (s *server) send(scalingReq *scalingRequest) error {
	if scalingReq == nil {
		return nil
	}
	ctx := context.Background()

	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, s.coreAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	req := request.NewScalingServiceClient(conn)
	var f func(ctx context.Context, in *request.ScalingRequest, opts ...grpc.CallOption) (*request.ScalingResponse, error)

	switch scalingReq.requestType {
	case "up":
		f = req.Up
	case "down":
		f = req.Down
	default:
		return fmt.Errorf("invalid request type: %s", scalingReq.requestType)
	}
	res, err := f(ctx, &request.ScalingRequest{
		Source:            scalingReq.source,
		Action:            scalingReq.action,
		ResourceGroupName: scalingReq.groupName,
	})
	if err != nil {
		return err
	}
	fmt.Printf("status: %s, job-id: %s\n", res.Status, res.ScalingJobId)
	return nil
}

type scalingRequest struct {
	source      string
	action      string
	groupName   string
	requestType string
}
