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

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: handler.proto

package handler

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// HandleServiceClient is the client API for HandleService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HandleServiceClient interface {
	// リソース操作の前イベント
	PreHandle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (HandleService_PreHandleClient, error)
	// リソース操作
	Handle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (HandleService_HandleClient, error)
	// リソース操作の後イベント
	PostHandle(ctx context.Context, in *PostHandleRequest, opts ...grpc.CallOption) (HandleService_PostHandleClient, error)
}

type handleServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHandleServiceClient(cc grpc.ClientConnInterface) HandleServiceClient {
	return &handleServiceClient{cc}
}

func (c *handleServiceClient) PreHandle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (HandleService_PreHandleClient, error) {
	stream, err := c.cc.NewStream(ctx, &HandleService_ServiceDesc.Streams[0], "/autoscaler.HandleService/PreHandle", opts...)
	if err != nil {
		return nil, err
	}
	x := &handleServicePreHandleClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HandleService_PreHandleClient interface {
	Recv() (*HandleResponse, error)
	grpc.ClientStream
}

type handleServicePreHandleClient struct {
	grpc.ClientStream
}

func (x *handleServicePreHandleClient) Recv() (*HandleResponse, error) {
	m := new(HandleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *handleServiceClient) Handle(ctx context.Context, in *HandleRequest, opts ...grpc.CallOption) (HandleService_HandleClient, error) {
	stream, err := c.cc.NewStream(ctx, &HandleService_ServiceDesc.Streams[1], "/autoscaler.HandleService/Handle", opts...)
	if err != nil {
		return nil, err
	}
	x := &handleServiceHandleClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HandleService_HandleClient interface {
	Recv() (*HandleResponse, error)
	grpc.ClientStream
}

type handleServiceHandleClient struct {
	grpc.ClientStream
}

func (x *handleServiceHandleClient) Recv() (*HandleResponse, error) {
	m := new(HandleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *handleServiceClient) PostHandle(ctx context.Context, in *PostHandleRequest, opts ...grpc.CallOption) (HandleService_PostHandleClient, error) {
	stream, err := c.cc.NewStream(ctx, &HandleService_ServiceDesc.Streams[2], "/autoscaler.HandleService/PostHandle", opts...)
	if err != nil {
		return nil, err
	}
	x := &handleServicePostHandleClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HandleService_PostHandleClient interface {
	Recv() (*HandleResponse, error)
	grpc.ClientStream
}

type handleServicePostHandleClient struct {
	grpc.ClientStream
}

func (x *handleServicePostHandleClient) Recv() (*HandleResponse, error) {
	m := new(HandleResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// HandleServiceServer is the server API for HandleService service.
// All implementations must embed UnimplementedHandleServiceServer
// for forward compatibility
type HandleServiceServer interface {
	// リソース操作の前イベント
	PreHandle(*HandleRequest, HandleService_PreHandleServer) error
	// リソース操作
	Handle(*HandleRequest, HandleService_HandleServer) error
	// リソース操作の後イベント
	PostHandle(*PostHandleRequest, HandleService_PostHandleServer) error
	mustEmbedUnimplementedHandleServiceServer()
}

// UnimplementedHandleServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHandleServiceServer struct {
}

func (UnimplementedHandleServiceServer) PreHandle(*HandleRequest, HandleService_PreHandleServer) error {
	return status.Errorf(codes.Unimplemented, "method PreHandle not implemented")
}
func (UnimplementedHandleServiceServer) Handle(*HandleRequest, HandleService_HandleServer) error {
	return status.Errorf(codes.Unimplemented, "method Handle not implemented")
}
func (UnimplementedHandleServiceServer) PostHandle(*PostHandleRequest, HandleService_PostHandleServer) error {
	return status.Errorf(codes.Unimplemented, "method PostHandle not implemented")
}
func (UnimplementedHandleServiceServer) mustEmbedUnimplementedHandleServiceServer() {}

// UnsafeHandleServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HandleServiceServer will
// result in compilation errors.
type UnsafeHandleServiceServer interface {
	mustEmbedUnimplementedHandleServiceServer()
}

func RegisterHandleServiceServer(s grpc.ServiceRegistrar, srv HandleServiceServer) {
	s.RegisterService(&HandleService_ServiceDesc, srv)
}

func _HandleService_PreHandle_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HandleRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HandleServiceServer).PreHandle(m, &handleServicePreHandleServer{stream})
}

type HandleService_PreHandleServer interface {
	Send(*HandleResponse) error
	grpc.ServerStream
}

type handleServicePreHandleServer struct {
	grpc.ServerStream
}

func (x *handleServicePreHandleServer) Send(m *HandleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _HandleService_Handle_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HandleRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HandleServiceServer).Handle(m, &handleServiceHandleServer{stream})
}

type HandleService_HandleServer interface {
	Send(*HandleResponse) error
	grpc.ServerStream
}

type handleServiceHandleServer struct {
	grpc.ServerStream
}

func (x *handleServiceHandleServer) Send(m *HandleResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _HandleService_PostHandle_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PostHandleRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HandleServiceServer).PostHandle(m, &handleServicePostHandleServer{stream})
}

type HandleService_PostHandleServer interface {
	Send(*HandleResponse) error
	grpc.ServerStream
}

type handleServicePostHandleServer struct {
	grpc.ServerStream
}

func (x *handleServicePostHandleServer) Send(m *HandleResponse) error {
	return x.ServerStream.SendMsg(m)
}

// HandleService_ServiceDesc is the grpc.ServiceDesc for HandleService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HandleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "autoscaler.HandleService",
	HandlerType: (*HandleServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PreHandle",
			Handler:       _HandleService_PreHandle_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Handle",
			Handler:       _HandleService_Handle_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PostHandle",
			Handler:       _HandleService_PostHandle_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "handler.proto",
}
