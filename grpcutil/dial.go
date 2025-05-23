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

package grpcutil

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sacloud/autoscaler/defaults"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DialOption struct {
	Destination          string
	TransportCredentials credentials.TransportCredentials
	DialOpts             []grpc.DialOption
}

func (opt *DialOption) destination() string {
	if opt.Destination != "" {
		return opt.Destination
	}
	for _, dest := range defaults.CoreSocketAddrCandidates {
		_, endpoint, err := parseTarget(dest)
		if err != nil {
			panic(err) // defaultsでの定義誤り
		}
		if _, err := os.Stat(endpoint); err == nil {
			return dest
		}
	}
	return ""
}

// DialContext 指定のオプションでgRPCクライアント接続を行い、コネクションとクリーンアップ用funcを返す
func DialContext(ctx context.Context, opt *DialOption) (*grpc.ClientConn, func(), error) {
	dest := opt.destination()
	if dest == "" {
		return nil, nil, fmt.Errorf(
			"default socket file not found in [%s]",
			strings.Join(defaults.CoreSocketAddrCandidates, ", "),
		)
	}

	var dialOpts []grpc.DialOption
	if opt.TransportCredentials != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opt.TransportCredentials))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	dialOpts = append(dialOpts, opt.DialOpts...)

	conn, err := grpc.DialContext(ctx, dest, dialOpts...)
	if err != nil {
		return nil, nil, err
	}
	return conn, func() {
		conn.Close()
	}, nil
}
