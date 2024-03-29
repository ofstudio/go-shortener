// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: api/internal.proto

package proto

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

// InternalClient is the client API for Internal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InternalClient interface {
	Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
}

type internalClient struct {
	cc grpc.ClientConnInterface
}

func NewInternalClient(cc grpc.ClientConnInterface) InternalClient {
	return &internalClient{cc}
}

func (c *internalClient) Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, "/proto.Internal/Stats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InternalServer is the server API for Internal service.
// All implementations must embed UnimplementedInternalServer
// for forward compatibility
type InternalServer interface {
	Stats(context.Context, *StatsRequest) (*StatsResponse, error)
	mustEmbedUnimplementedInternalServer()
}

// UnimplementedInternalServer must be embedded to have forward compatible implementations.
type UnimplementedInternalServer struct {
}

func (UnimplementedInternalServer) Stats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
func (UnimplementedInternalServer) mustEmbedUnimplementedInternalServer() {}

// UnsafeInternalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InternalServer will
// result in compilation errors.
type UnsafeInternalServer interface {
	mustEmbedUnimplementedInternalServer()
}

func RegisterInternalServer(s grpc.ServiceRegistrar, srv InternalServer) {
	s.RegisterService(&Internal_ServiceDesc, srv)
}

func _Internal_Stats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InternalServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Internal/Stats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InternalServer).Stats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Internal_ServiceDesc is the grpc.ServiceDesc for Internal service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Internal_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Internal",
	HandlerType: (*InternalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Stats",
			Handler:    _Internal_Stats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/internal.proto",
}
