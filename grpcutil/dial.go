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
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DialOption struct {
	Destination          string
	TransportCredentials credentials.TransportCredentials
}

// DialContext 指定のオプションでgRPCクライアント接続を行い、コネクションとクリーンアップ用funcを返す
func DialContext(ctx context.Context, opt *DialOption) (*grpc.ClientConn, func(), error) {
	var dialOpts []grpc.DialOption
	if opt.TransportCredentials != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opt.TransportCredentials))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, opt.Destination, dialOpts...)
	if err != nil {
		return nil, nil, err
	}
	return conn, func() {
		conn.Close() // nolint
	}, nil
}
