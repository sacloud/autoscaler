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

package grpcutil

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/sacloud/autoscaler/config"
	"google.golang.org/grpc"
)

type ListenerOption struct {
	Address    string
	TLSConfig  *config.TLSStruct
	ServerOpts []grpc.ServerOption
}

// Server 指定のオプションでリッスン構成をした後でリッスンし、*grpc.Serverとクリーンアップ用のfuncを返す
func Server(opt *ListenerOption) (*grpc.Server, net.Listener, func(), error) {
	schema, endpoint, err := ParseTarget(opt.Address)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ParseTarget failed: %s", err)
	}

	listener, err := net.Listen(schema, endpoint)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("net.Listen failed: %s", err)
	}

	cleanup := func() {
		listener.Close() // nolint
		if schema == "unix" {
			if _, err := os.Stat(endpoint); err == nil {
				if err := os.RemoveAll(endpoint); err != nil {
					log.Printf("cleanup failed: %s", err) // nolint
				}
			}
		}
	}

	var serverOpts []grpc.ServerOption
	if opt.TLSConfig != nil && schema != "unix" {
		cred, err := opt.TLSConfig.TransportCredentials()
		if err != nil && err != config.ErrNoTLSConfig {
			return nil, nil, nil, err
		}
		serverOpts = append(serverOpts, grpc.Creds(cred))
	}
	serverOpts = append(serverOpts, opt.ServerOpts...)

	return grpc.NewServer(serverOpts...), listener, cleanup, nil
}
