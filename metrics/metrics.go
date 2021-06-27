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

package metrics

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/log"
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
	TLSConfig     *config.TLSStruct

	logger *log.Logger
	server *http.Server
}

func NewServer(addr string, tlsConfig *config.TLSStruct, logger *log.Logger) *Server {
	return &Server{
		ListenAddress: addr,
		TLSConfig:     tlsConfig,
		logger:        logger,
		server:        &http.Server{Addr: addr, Handler: handler()},
	}
}

func (s *Server) Serve(listener net.Listener) error {
	if err := s.logger.Info("message", "exporter started", "address", listener.Addr().String()); err != nil {
		return err
	}

	if s.TLSConfig != nil {
		conf, err := s.TLSConfig.TLSConfig()
		if err != nil {
			return err
		}
		s.server.TLSConfig = conf
		if err := s.logger.Info("message", "exporter has enabled TLS"); err != nil {
			return err
		}
		return s.server.ServeTLS(listener, "", "")
	}

	return s.server.Serve(listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
