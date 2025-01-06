// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

func TestServe(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	server := NewServer(listener.Addr().String(), test.Logger)
	closed := make(chan struct{})
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			t.Error(err)
		}
		close(closed)
	}()

	res, err := http.Get("http://" + listener.Addr().String() + "/metrics")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	require.True(t, strings.Contains(string(data), "promhttp_metric_handler_requests_total"))

	if err := server.server.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	<-closed
}
