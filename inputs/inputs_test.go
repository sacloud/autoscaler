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

package inputs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/prometheus/common/expfmt"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/metrics"
	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

func init() {
	initMetrics()
}

type fakeInput struct {
	listenAddr string
	configPath string
}

func (i *fakeInput) Name() string {
	return "fake"
}
func (i *fakeInput) Version() string {
	return "dev"
}
func (i *fakeInput) ShouldAccept(req *http.Request) (bool, error) {
	return false, nil
}
func (i *fakeInput) Destination() string {
	return ""
}
func (i *fakeInput) ListenAddress() string {
	return i.listenAddr
}
func (i *fakeInput) ConfigPath() string {
	return i.configPath
}
func (i *fakeInput) GetLogger() *log.Logger {
	return test.Logger
}

func Test_server_serve(t *testing.T) {
	tests := []struct {
		name           string
		schema         string
		configPath     string
		clientKeyPath  string
		clientCertPath string
		caCertPath     string
		statusCode     int
		wantErr        bool
		forceHTTP2     bool
	}{
		{
			name:       "http without ConfigPath",
			schema:     "http",
			statusCode: http.StatusOK,
		},
		{
			name:    "https without ConfigPath",
			schema:  "https",
			wantErr: true,
		},
		{
			name:       "https with minimal Config",
			schema:     "https",
			configPath: "./test/inputs.minimal.yaml",
			statusCode: http.StatusOK,
		},
		{
			name:       "http with minimal Config",
			schema:     "http",
			configPath: "./test/inputs.minimal.yaml",
			statusCode: http.StatusBadRequest,
		},
		{
			name:           "with mtls Config",
			schema:         "https",
			configPath:     "./test/inputs.mtls.yaml",
			clientCertPath: "./test/client-cert.pem",
			clientKeyPath:  "./test/client-key.pem",
			caCertPath:     "./test/ca-cert.pem",
			statusCode:     http.StatusOK,
		},
		{
			name:       "with mtls Config without client cert",
			schema:     "http",
			configPath: "./test/inputs.mtls.yaml",
			caCertPath: "./test/ca-cert.pem",
			statusCode: http.StatusBadRequest,
		},
		{
			name:           "with mtls and HTTP/2",
			schema:         "https",
			configPath:     "./test/inputs.mtls.yaml",
			clientCertPath: "./test/client-cert.pem",
			clientKeyPath:  "./test/client-key.pem",
			caCertPath:     "./test/ca-cert.pem",
			statusCode:     http.StatusOK,
			forceHTTP2:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &fakeInput{
				listenAddr: "localhost:0",
				configPath: tt.configPath,
			}
			conf, err := LoadConfigFromPath(tt.configPath)
			if err != nil {
				t.Fatal(err)
			}
			server, err := newServer(input, conf)
			if err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatal(err)
			}

			closed := make(chan struct{})
			go func() {
				if err := server.serve(listener); err != http.ErrServerClosed {
					t.Log(err)
				}
				close(closed)
			}()

			client := testHTTPClient(t, tt.clientKeyPath, tt.clientCertPath, tt.caCertPath, tt.forceHTTP2)
			url := fmt.Sprintf("%s://%s/healthz", tt.schema, listener.Addr().String())
			res, err := client.Get(url)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if tt.forceHTTP2 {
					require.Equal(t, "HTTP/2.0", res.Proto)
				}
				require.Equal(t, tt.statusCode, res.StatusCode)
				if res.StatusCode == http.StatusOK {
					data, err := io.ReadAll(res.Body)
					if err != nil {
						t.Fatal(err)
					}
					require.Equal(t, "ok", string(data))
				}
			}

			if err := server.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			<-closed
		})
	}
}

func Test_server_exporter(t *testing.T) {
	input := &fakeInput{
		listenAddr: "localhost:0",
	}

	closed1 := make(chan struct{})
	closed2 := make(chan struct{})
	// inputs server
	server, err := newServer(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := server.serve(listener); err != http.ErrServerClosed {
			t.Log(err)
		}
		close(closed1)
	}()

	// exporter
	exporterListener, err := net.Listen("tcp", "localhost:0")
	exporterServer := metrics.NewServer(exporterListener.Addr().String(), nil, input.GetLogger())
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := exporterServer.Serve(exporterListener); err != http.ErrServerClosed {
			t.Log(err)
		}
		close(closed2)
	}()

	upURL := fmt.Sprintf("http://%s/up", listener.Addr().String())
	downURL := fmt.Sprintf("http://%s/down", listener.Addr().String())
	metricsURL := fmt.Sprintf("http://%s/metrics", exporterListener.Addr().String())

	// /up * 2、 /down * 1
	requests := []string{upURL, upURL, downURL}
	for _, url := range requests {
		if _, err := http.Get(url); err != nil {
			t.Fatal(err)
		}
	}

	req, err := http.Get(metricsURL)
	if err != nil {
		t.Fatal(err)
	}

	// parse response of Exporter
	var parser expfmt.TextParser
	parsed, err := parser.TextToMetricFamilies(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	reqTotal := parsed["sacloud_autoscaler_webhook_requests_total"]
	upTotal := parsed["sacloud_autoscaler_webhook_requests_up"]
	downTotal := parsed["sacloud_autoscaler_webhook_requests_down"]

	require.NotNil(t, reqTotal)
	require.NotNil(t, upTotal)
	require.NotNil(t, downTotal)

	totalMetrics := reqTotal.GetMetric()
	require.Len(t, totalMetrics, 3)
	require.Equal(t, totalMetrics[0].GetLabel()[0].GetValue(), "200")
	require.Equal(t, totalMetrics[0].GetCounter().GetValue(), float64(3))
	require.Equal(t, totalMetrics[1].GetLabel()[0].GetValue(), "400")
	require.Equal(t, totalMetrics[1].GetCounter().GetValue(), float64(0))
	require.Equal(t, totalMetrics[2].GetLabel()[0].GetValue(), "500")
	require.Equal(t, totalMetrics[2].GetCounter().GetValue(), float64(0))

	upMetrics := upTotal.GetMetric()
	require.Len(t, upMetrics, 3)
	require.Equal(t, upMetrics[0].GetLabel()[0].GetValue(), "200")
	require.Equal(t, upMetrics[0].GetCounter().GetValue(), float64(2))
	require.Equal(t, upMetrics[1].GetLabel()[0].GetValue(), "400")
	require.Equal(t, upMetrics[1].GetCounter().GetValue(), float64(0))
	require.Equal(t, upMetrics[2].GetLabel()[0].GetValue(), "500")
	require.Equal(t, upMetrics[2].GetCounter().GetValue(), float64(0))

	downMetrics := downTotal.GetMetric()
	require.Len(t, downMetrics, 3)
	require.Equal(t, downMetrics[0].GetLabel()[0].GetValue(), "200")
	require.Equal(t, downMetrics[0].GetCounter().GetValue(), float64(1))
	require.Equal(t, downMetrics[1].GetLabel()[0].GetValue(), "400")
	require.Equal(t, downMetrics[1].GetCounter().GetValue(), float64(0))
	require.Equal(t, downMetrics[2].GetLabel()[0].GetValue(), "500")
	require.Equal(t, downMetrics[2].GetCounter().GetValue(), float64(0))

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := exporterServer.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-closed1
	<-closed2
}
func testHTTPClient(t *testing.T, clientKeyPath, clientCertPath, caCertPath string, forceHTTP2 bool) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if clientKeyPath != "" && clientCertPath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			t.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	var transport http.RoundTripper = &http.Transport{
		TLSClientConfig:   tlsConfig,
		ForceAttemptHTTP2: forceHTTP2,
	}
	return &http.Client{Transport: transport}
}
