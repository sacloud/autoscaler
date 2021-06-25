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

package grpcutil

import (
	"fmt"
	"log"
	"net"
	"os"
)

type ListenerOption struct {
	Address string
}

// Listener 指定のオプションでリッスン構成をした後でリッスンし、net.Listenerとクリーンアップ用のfuncを返す
func Listener(opt *ListenerOption) (net.Listener, func(), error) {
	target := ParseTarget(opt.Address, false)

	schema := "tcp"
	if target.Scheme == "unix" || target.Scheme == "unix-abstract" {
		schema = "unix"
	}

	listener, err := net.Listen(schema, target.Endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("net.Listen failed: %s", err)
	}

	return listener, func() {
		listener.Close() // nolint
		if schema == "unix" {
			if _, err := os.Stat(target.Endpoint); err == nil {
				if err := os.RemoveAll(target.Endpoint); err != nil {
					log.Printf("cleanup failed: %s", err) // nolint
				}
			}
		}
	}, nil
}
