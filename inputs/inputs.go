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
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
)

var webhookBodyMaxLen = int64(64 * 1024) // 64KB

type Input interface {
	Name() string
	Version() string
	ShouldAccept(req *http.Request) (bool, error) // true,nilを返した場合のみCoreへのリクエストを行う
	Destination() string
	ListenAddress() string
	GetLogger() *log.Logger
}

func FullName(input Input) string {
	return fmt.Sprintf("autoscaler-inputs-%s", input.Name())
}

func Serve(input Input) error {
	server := &server{
		coreAddress:   input.Destination(),
		listenAddress: input.ListenAddress(),
		input:         input,
		logger:        input.GetLogger().WithPrefix("from", FullName(input)),
	}
	return server.listenAndServe()
}

type server struct {
	coreAddress   string
	listenAddress string
	input         Input
	logger        *log.Logger
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

	if err := s.logger.Info("message", "started", "address", s.listenAddress); err != nil {
		return err
	}

	return http.ListenAndServe(s.listenAddress, serveMux)
}

func (s *server) handle(requestType string, w http.ResponseWriter, req *http.Request) {
	// bodyをwebhookBodyMaxLenまでに制限
	req.Body = http.MaxBytesReader(w, req.Body, webhookBodyMaxLen)

	scalingReq, err := s.parseRequest(requestType, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error())) // nolint
		return
	}
	if scalingReq == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ignored")) // nolint
		return
	}

	if err := s.send(scalingReq); err != nil {
		s.logger.Error("error", err) // nolint
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("accepted")) // nolint
}

func (s *server) parseRequest(requestType string, req *http.Request) (*ScalingRequest, error) {
	if err := s.logger.Info("message", "webhook received"); err != nil {
		return nil, err
	}

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	if err := s.logger.Debug("request", string(dump)); err != nil {
		return nil, err
	}

	shouldAccept, err := s.input.ShouldAccept(req)
	if err != nil {
		return nil, err
	}
	if !shouldAccept {
		if err := s.logger.Info("message", "webhook ignored"); err != nil {
			return nil, err
		}
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
	groupName := queryStrings.Get("resource-group-name")
	if groupName == "" {
		groupName = defaults.ResourceGroupName
	}
	desiredStateName := queryStrings.Get("desired-state-name")
	if desiredStateName == "" {
		desiredStateName = defaults.DesiredStateName
	}

	scalingReq := &ScalingRequest{
		Source:           source,
		Action:           action,
		GroupName:        groupName,
		RequestType:      requestType,
		DesiredStateName: desiredStateName,
	}
	if err := scalingReq.Validate(); err != nil {
		return nil, err
	}
	return scalingReq, nil
}

func (s *server) send(scalingReq *ScalingRequest) error {
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

	switch scalingReq.RequestType {
	case "up":
		f = req.Up
	case "down":
		f = req.Down
	default:
		return fmt.Errorf("invalid request type: %s", scalingReq.RequestType)
	}
	res, err := f(ctx, &request.ScalingRequest{
		Source:            scalingReq.Source,
		Action:            scalingReq.Action,
		ResourceGroupName: scalingReq.GroupName,
		DesiredStateName:  scalingReq.DesiredStateName,
	})
	if err != nil {
		return err
	}
	return s.logger.Info("message", "webhook handled", "status", res.Status, "job-id", res.ScalingJobId)
}
