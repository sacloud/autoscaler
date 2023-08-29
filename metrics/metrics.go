// Copyright 2021-2023 The sacloud/autoscaler Authors
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

package metrics

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

var errors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "sacloud_autoscaler_grpc_errors_total",
	Help: "The total number of errors",
}, []string{"component"})

// IncrementErrorCount sacloud_autoscaler_grpc_errors_totalを指定のラベルとともにインクリメントする
func IncrementErrorCount(component string) {
	errors.WithLabelValues(component).Inc()
}

func InitErrorCount(component string) {
	errors.WithLabelValues(component)
}

// Server メトリクス収集用の*http.Serverラッパー
type Server struct {
	ListenAddress string

	logger *slog.Logger
	server *http.Server
}

func NewServer(addr string, logger *slog.Logger) *Server {
	return &Server{
		ListenAddress: addr,
		logger:        logger,
		server:        &http.Server{Addr: addr, Handler: handler()}, //nolint
	}
}

func (s *Server) Serve(listener net.Listener) error {
	s.logger.Info("exporter started", slog.String("address", listener.Addr().String()))

	return s.server.Serve(listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
