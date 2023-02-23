// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: test-services.proto

package pbsuite

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

// HelloServiceClient is the client API for HelloService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HelloServiceClient interface {
	Hello(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*HelloResponse, error)
}

type helloServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHelloServiceClient(cc grpc.ClientConnInterface) HelloServiceClient {
	return &helloServiceClient{cc}
}

func (c *helloServiceClient) Hello(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := c.cc.Invoke(ctx, "/pbsuite.HelloService/Hello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HelloServiceServer is the server API for HelloService service.
// All implementations must embed UnimplementedHelloServiceServer
// for forward compatibility
type HelloServiceServer interface {
	Hello(context.Context, *Empty) (*HelloResponse, error)
	mustEmbedUnimplementedHelloServiceServer()
}

// UnimplementedHelloServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHelloServiceServer struct {
}

func (UnimplementedHelloServiceServer) Hello(context.Context, *Empty) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedHelloServiceServer) mustEmbedUnimplementedHelloServiceServer() {}

// UnsafeHelloServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HelloServiceServer will
// result in compilation errors.
type UnsafeHelloServiceServer interface {
	mustEmbedUnimplementedHelloServiceServer()
}

func RegisterHelloServiceServer(s grpc.ServiceRegistrar, srv HelloServiceServer) {
	s.RegisterService(&HelloService_ServiceDesc, srv)
}

func _HelloService_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServiceServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pbsuite.HelloService/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServiceServer).Hello(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// HelloService_ServiceDesc is the grpc.ServiceDesc for HelloService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HelloService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pbsuite.HelloService",
	HandlerType: (*HelloServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _HelloService_Hello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test-services.proto",
}

// AnswerServiceClient is the client API for AnswerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AnswerServiceClient interface {
	Answer(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*AnswerResponse, error)
}

type answerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAnswerServiceClient(cc grpc.ClientConnInterface) AnswerServiceClient {
	return &answerServiceClient{cc}
}

func (c *answerServiceClient) Answer(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*AnswerResponse, error) {
	out := new(AnswerResponse)
	err := c.cc.Invoke(ctx, "/pbsuite.AnswerService/Answer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AnswerServiceServer is the server API for AnswerService service.
// All implementations must embed UnimplementedAnswerServiceServer
// for forward compatibility
type AnswerServiceServer interface {
	Answer(context.Context, *Empty) (*AnswerResponse, error)
	mustEmbedUnimplementedAnswerServiceServer()
}

// UnimplementedAnswerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAnswerServiceServer struct {
}

func (UnimplementedAnswerServiceServer) Answer(context.Context, *Empty) (*AnswerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Answer not implemented")
}
func (UnimplementedAnswerServiceServer) mustEmbedUnimplementedAnswerServiceServer() {}

// UnsafeAnswerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AnswerServiceServer will
// result in compilation errors.
type UnsafeAnswerServiceServer interface {
	mustEmbedUnimplementedAnswerServiceServer()
}

func RegisterAnswerServiceServer(s grpc.ServiceRegistrar, srv AnswerServiceServer) {
	s.RegisterService(&AnswerService_ServiceDesc, srv)
}

func _AnswerService_Answer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnswerServiceServer).Answer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pbsuite.AnswerService/Answer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnswerServiceServer).Answer(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// AnswerService_ServiceDesc is the grpc.ServiceDesc for AnswerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AnswerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pbsuite.AnswerService",
	HandlerType: (*AnswerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Answer",
			Handler:    _AnswerService_Answer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test-services.proto",
}