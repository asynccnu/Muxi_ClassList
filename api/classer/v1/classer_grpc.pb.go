// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: classer/v1/classer.proto

package v1

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

const (
	Classer_GetClass_FullMethodName    = "/classer.v1.Classer/GetClass"
	Classer_AddClass_FullMethodName    = "/classer.v1.Classer/AddClass"
	Classer_DeleteClass_FullMethodName = "/classer.v1.Classer/DeleteClass"
	Classer_UpdateClass_FullMethodName = "/classer.v1.Classer/UpdateClass"
)

// ClasserClient is the client API for Classer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClasserClient interface {
	// Sends a greeting
	GetClass(ctx context.Context, in *GetClassRequest, opts ...grpc.CallOption) (*GetClassResponse, error)
	AddClass(ctx context.Context, in *AddClassRequest, opts ...grpc.CallOption) (*AddClassResponse, error)
	DeleteClass(ctx context.Context, in *DeleteClassRequest, opts ...grpc.CallOption) (*DeleteClassResponse, error)
	UpdateClass(ctx context.Context, in *UpdateClassRequest, opts ...grpc.CallOption) (*UpdateClassResponse, error)
}

type classerClient struct {
	cc grpc.ClientConnInterface
}

func NewClasserClient(cc grpc.ClientConnInterface) ClasserClient {
	return &classerClient{cc}
}

func (c *classerClient) GetClass(ctx context.Context, in *GetClassRequest, opts ...grpc.CallOption) (*GetClassResponse, error) {
	out := new(GetClassResponse)
	err := c.cc.Invoke(ctx, Classer_GetClass_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *classerClient) AddClass(ctx context.Context, in *AddClassRequest, opts ...grpc.CallOption) (*AddClassResponse, error) {
	out := new(AddClassResponse)
	err := c.cc.Invoke(ctx, Classer_AddClass_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *classerClient) DeleteClass(ctx context.Context, in *DeleteClassRequest, opts ...grpc.CallOption) (*DeleteClassResponse, error) {
	out := new(DeleteClassResponse)
	err := c.cc.Invoke(ctx, Classer_DeleteClass_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *classerClient) UpdateClass(ctx context.Context, in *UpdateClassRequest, opts ...grpc.CallOption) (*UpdateClassResponse, error) {
	out := new(UpdateClassResponse)
	err := c.cc.Invoke(ctx, Classer_UpdateClass_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClasserServer is the server API for Classer service.
// All implementations must embed UnimplementedClasserServer
// for forward compatibility
type ClasserServer interface {
	// Sends a greeting
	GetClass(context.Context, *GetClassRequest) (*GetClassResponse, error)
	AddClass(context.Context, *AddClassRequest) (*AddClassResponse, error)
	DeleteClass(context.Context, *DeleteClassRequest) (*DeleteClassResponse, error)
	UpdateClass(context.Context, *UpdateClassRequest) (*UpdateClassResponse, error)
	mustEmbedUnimplementedClasserServer()
}

// UnimplementedClasserServer must be embedded to have forward compatible implementations.
type UnimplementedClasserServer struct {
}

func (UnimplementedClasserServer) GetClass(context.Context, *GetClassRequest) (*GetClassResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClass not implemented")
}
func (UnimplementedClasserServer) AddClass(context.Context, *AddClassRequest) (*AddClassResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddClass not implemented")
}
func (UnimplementedClasserServer) DeleteClass(context.Context, *DeleteClassRequest) (*DeleteClassResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteClass not implemented")
}
func (UnimplementedClasserServer) UpdateClass(context.Context, *UpdateClassRequest) (*UpdateClassResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateClass not implemented")
}
func (UnimplementedClasserServer) mustEmbedUnimplementedClasserServer() {}

// UnsafeClasserServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClasserServer will
// result in compilation errors.
type UnsafeClasserServer interface {
	mustEmbedUnimplementedClasserServer()
}

func RegisterClasserServer(s grpc.ServiceRegistrar, srv ClasserServer) {
	s.RegisterService(&Classer_ServiceDesc, srv)
}

func _Classer_GetClass_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClassRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClasserServer).GetClass(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Classer_GetClass_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClasserServer).GetClass(ctx, req.(*GetClassRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Classer_AddClass_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddClassRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClasserServer).AddClass(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Classer_AddClass_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClasserServer).AddClass(ctx, req.(*AddClassRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Classer_DeleteClass_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteClassRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClasserServer).DeleteClass(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Classer_DeleteClass_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClasserServer).DeleteClass(ctx, req.(*DeleteClassRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Classer_UpdateClass_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateClassRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClasserServer).UpdateClass(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Classer_UpdateClass_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClasserServer).UpdateClass(ctx, req.(*UpdateClassRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Classer_ServiceDesc is the grpc.ServiceDesc for Classer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Classer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "classer.v1.Classer",
	HandlerType: (*ClasserServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetClass",
			Handler:    _Classer_GetClass_Handler,
		},
		{
			MethodName: "AddClass",
			Handler:    _Classer_AddClass_Handler,
		},
		{
			MethodName: "DeleteClass",
			Handler:    _Classer_DeleteClass_Handler,
		},
		{
			MethodName: "UpdateClass",
			Handler:    _Classer_UpdateClass_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "classer/v1/classer.proto",
}
