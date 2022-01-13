// This file copied from github.com/grpc/grpc-go.
// Original License is as follows:

/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package grpcutil provides a bunch of utility functions to be used across the
// gRPC codebase.
package grpcutil

import (
	"net/url"
	"strings"
)

// ParseTarget gRPCエンドポイントアドレス文字列を受け取り、スキーマ/エンドポイントをパースして返す
func ParseTarget(target string) (string, string, error) {
	if !strings.HasPrefix(target, "unix:") && !strings.HasPrefix(target, "unix-abstract:") {
		target = "tcp:" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", "", err
	}

	schema := "tcp"
	if u.Scheme == "unix" || u.Scheme == "unix-abstract" {
		schema = "unix"
	}

	endpoint := u.Path
	if endpoint == "" {
		endpoint = u.Opaque
	}
	return schema, endpoint, nil
}
