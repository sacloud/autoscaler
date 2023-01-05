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

package grpcutil

import (
	"context"

	"github.com/sacloud/autoscaler/metrics"
	"google.golang.org/grpc"
)

func ServerErrorCountInterceptor(component string) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			res, err := handler(ctx, req)
			if err != nil {
				metrics.IncrementErrorCount(component)
			}
			return res, err
		}),
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			err := handler(srv, ss)
			if err != nil {
				metrics.IncrementErrorCount(component)
			}
			return err
		}),
	}
}

func ClientErrorCountInterceptor(component string) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err != nil {
				metrics.IncrementErrorCount(component)
			}
			return err
		}),
		grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			ret, err := streamer(ctx, desc, cc, method, opts...)
			if err != nil {
				metrics.IncrementErrorCount(component)
			}
			return ret, err
		}),
	}
}
