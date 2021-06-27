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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

type fakeInput struct {
	listenAddr    string
	tlsConfigPath string
}

func (i *fakeInput) Name() string {
	return "fake"
}
func (i *fakeInput) Version() string {
	return "dev"
}
func (i *fakeInput) ShouldAccept(req *http.Request) (bool, error) {
	return true, nil
}
func (i *fakeInput) Destination() string {
	return ""
}
func (i *fakeInput) ListenAddress() string {
	return i.listenAddr
}
func (i *fakeInput) TLSConfigPath() string {
	return i.tlsConfigPath
}
func (i *fakeInput) GetLogger() *log.Logger {
	return test.Logger
}

func Test_server_serve(t *testing.T) {
	tests := []struct {
		name           string
		schema         string
		webConfigPath  string
		clientKeyPath  string
		clientCertPath string
		caCertPath     string
		statusCode     int
		wantErr        bool
		forceHTTP2     bool
	}{
		{
			name:       "http without TLSConfigPath",
			schema:     "http",
			statusCode: http.StatusOK,
		},
		{
			name:    "https without TLSConfigPath",
			schema:  "https",
			wantErr: true,
		},
		{
			name:          "https with minimal TLSConfig",
			schema:        "https",
			webConfigPath: "../test/inputs.minimal.yaml",
			statusCode:    http.StatusOK,
		},
		{
			name:          "http with minimal TLSConfig",
			schema:        "http",
			webConfigPath: "../test/inputs.minimal.yaml",
			statusCode:    http.StatusBadRequest,
		},
		{
			name:           "with mtls TLSConfig",
			schema:         "https",
			webConfigPath:  "../test/inputs.mtls.yaml",
			clientCertPath: "../test/client-cert.pem",
			clientKeyPath:  "../test/client-key.pem",
			caCertPath:     "../test/ca-cert.pem",
			statusCode:     http.StatusOK,
		},
		{
			name:          "with mtls TLSConfig without client cert",
			schema:        "http",
			webConfigPath: "../test/inputs.mtls.yaml",
			caCertPath:    "../test/ca-cert.pem",
			statusCode:    http.StatusBadRequest,
		},
		{
			name:           "with mtls and HTTP/2",
			schema:         "https",
			webConfigPath:  "../test/inputs.mtls.yaml",
			clientCertPath: "../test/client-cert.pem",
			clientKeyPath:  "../test/client-key.pem",
			caCertPath:     "../test/ca-cert.pem",
			statusCode:     http.StatusOK,
			forceHTTP2:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &fakeInput{
				listenAddr:    "localhost:0",
				tlsConfigPath: tt.webConfigPath,
			}
			server, err := newServer(input)
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
