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
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/metrics"
	"github.com/sacloud/autoscaler/request"
	"google.golang.org/grpc"
)

var (
	webhookBodyMaxLen      = int64(64 * 1024) // 64KB
	allowedQueryStringKeys = []string{
		"source", "action", "resource-group-name", "desired-state-name",
	}
)

// Input Webhookを受け取りCoreへのリクエストを行うInputsが備えるべきインターフェース
type Input interface {
	Name() string
	Version() string
	ShouldAccept(req *http.Request) (bool, error) // true,nilを返した場合のみCoreへのリクエストを行う
	Destination() string
	ListenAddress() string
	ConfigPath() string
	GetLogger() *log.Logger
}

func FullName(input Input) string {
	return fmt.Sprintf("autoscaler-inputs-%s", input.Name())
}

func Serve(ctx context.Context, input Input) error {
	initMetrics()

	errCh := make(chan error)

	conf, err := LoadConfigFromPath(input.ConfigPath())
	if err != nil {
		return err
	}

	// webhook
	go func() {
		errCh <- startWebhookServer(ctx, input, conf)
	}()

	// exporter
	if conf != nil && conf.ExporterConfig != nil && conf.ExporterConfig.Enabled {
		go func() {
			errCh <- startExporter(ctx, input, conf.ExporterConfig)
		}()
	}

	select {
	case err := <-errCh:
		return fmt.Errorf("inputs service failed: %s", err)
	case <-ctx.Done():
		input.GetLogger().Info("message", "shutting down", "error", ctx.Err()) // nolint
	}
	return ctx.Err()
}

func startWebhookServer(_ context.Context, input Input, conf *Config) error {
	server, err := newServer(input, conf)
	if err != nil {
		return err
	}
	return server.listenAndServe()
}

func startExporter(ctx context.Context, input Input, conf *config.ExporterConfig) error {
	if !conf.Enabled {
		return nil
	}
	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return err
	}
	server := metrics.NewServer(listener.Addr().String(), conf.TLSConfig, input.GetLogger())
	return server.Serve(listener)
}

type server struct {
	coreAddress   string
	listenAddress string
	webConfigPath string
	input         Input
	logger        *log.Logger
	config        *Config

	*http.Server
}

func newServer(input Input, conf *Config) (*server, error) {
	serveMux := http.NewServeMux()

	s := &server{
		coreAddress:   input.Destination(),
		listenAddress: input.ListenAddress(),
		webConfigPath: input.ConfigPath(),
		input:         input,
		logger:        input.GetLogger(),
		config:        conf,
		Server:        &http.Server{Addr: input.ListenAddress(), Handler: serveMux},
	}

	upWebhookHandler := promhttp.InstrumentHandlerCounter(
		counter,
		promhttp.InstrumentHandlerCounter(
			upCounter,
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				s.handle("up", w, req)
			}),
		),
	)
	downWebhookHandler := promhttp.InstrumentHandlerCounter(
		counter,
		promhttp.InstrumentHandlerCounter(
			downCounter,
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				s.handle("down", w, req)
			}),
		),
	)

	serveMux.HandleFunc("/up", upWebhookHandler)
	serveMux.HandleFunc("/down", downWebhookHandler)

	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) // nolint
	})

	return s, nil
}

func (s *server) listenAndServe() error {
	listener, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	defer listener.Close()

	return s.serve(listener)
}

func (s *server) serve(l net.Listener) error {
	if err := s.logger.Info("message", "started", "address", l.Addr().String()); err != nil {
		return err
	}

	if s.config != nil && s.config.ServerTLSConfig != nil {
		tlsConfig, err := s.config.ServerTLSConfig.TLSConfig()
		if err != nil {
			if err == config.ErrNoTLSConfig {
				return s.Serve(l)
			}
		}
		s.TLSConfig = tlsConfig
		return s.ServeTLS(l, "", "")
	}
	return s.Serve(l)
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
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ignored"}`)) // nolint
		return
	}

	s.logger.Info("message", "sending request to the Core server", "request-type", scalingReq.RequestType) // nolint

	res, err := s.send(scalingReq)
	if err != nil {
		s.logger.Error("error", err) // nolint
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.logger.Info(
		"message", "webhook handled",
		"status", res.Status,
		"job-id", res.ScalingJobId,
		"job-message", res.Message,
	); err != nil {
		s.logger.Error("error", err) // nolint
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id":"%s", "status":"%s", "message":"%s"}`, res.ScalingJobId, res.Status, res.Message))) // nolint
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
	if err := s.validateQueryString(queryStrings); err != nil {
		return nil, err
	}

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

func (s *server) validateQueryString(query url.Values) error {
	errors := &multierror.Error{}
	for k := range query {
		found := false
		for _, allowed := range allowedQueryStringKeys {
			if k == allowed {
				found = true
				break
			}
		}
		if !found {
			errors = multierror.Append(errors, fmt.Errorf("invalid parameter key: %s", k))
		}
	}
	if err := errors.ErrorOrNil(); err != nil {
		return err
	}
	return nil
}

func (s *server) send(scalingReq *ScalingRequest) (*request.ScalingResponse, error) {
	if scalingReq == nil {
		return nil, nil
	}
	ctx := context.Background()

	dialOption := &grpcutil.DialOption{
		Destination: s.coreAddress,
		DialOpts:    grpcutil.ClientErrorCountInterceptor("inputs_to_core"),
	}
	if s.config != nil && s.config.CoreTLSConfig != nil {
		cred, err := s.config.CoreTLSConfig.TransportCredentials()
		if err != nil && err != config.ErrNoTLSConfig {
			return nil, err
		}
		dialOption.TransportCredentials = cred
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, dialOption)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	req := request.NewScalingServiceClient(conn)
	var f func(ctx context.Context, in *request.ScalingRequest, opts ...grpc.CallOption) (*request.ScalingResponse, error)

	switch scalingReq.RequestType {
	case "up":
		f = req.Up
	case "down":
		f = req.Down
	default:
		return nil, fmt.Errorf("invalid request type: %s", scalingReq.RequestType)
	}
	return f(ctx, &request.ScalingRequest{
		Source:            scalingReq.Source,
		Action:            scalingReq.Action,
		ResourceGroupName: scalingReq.GroupName,
		DesiredStateName:  scalingReq.DesiredStateName,
	})
}
