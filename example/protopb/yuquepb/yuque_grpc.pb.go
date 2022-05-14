// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package yuquepb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// YuqueClient is the client API for Yuque service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type YuqueClient interface {
	// 获取认证的用户的个人信息
	UserInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserInfoResp, error)
	// 获取单个用户信息
	UserInfoByLogin(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error)
	// 创建 Group
	CreateGroup(ctx context.Context, in *CreateGroupReq, opts ...grpc.CallOption) (*CreateGroupResp, error)
}

type yuqueClient struct {
	cc grpc.ClientConnInterface
}

func NewYuqueClient(cc grpc.ClientConnInterface) YuqueClient {
	return &yuqueClient{cc}
}

func (c *yuqueClient) UserInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserInfoResp, error) {
	out := new(UserInfoResp)
	err := c.cc.Invoke(ctx, "/yuque.v2.Yuque/UserInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *yuqueClient) UserInfoByLogin(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error) {
	out := new(UserInfoResp)
	err := c.cc.Invoke(ctx, "/yuque.v2.Yuque/UserInfoByLogin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *yuqueClient) CreateGroup(ctx context.Context, in *CreateGroupReq, opts ...grpc.CallOption) (*CreateGroupResp, error) {
	out := new(CreateGroupResp)
	err := c.cc.Invoke(ctx, "/yuque.v2.Yuque/CreateGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// YuqueServer is the server API for Yuque service.
// All implementations should embed UnimplementedYuqueServer
// for forward compatibility
type YuqueServer interface {
	// 获取认证的用户的个人信息
	UserInfo(context.Context, *emptypb.Empty) (*UserInfoResp, error)
	// 获取单个用户信息
	UserInfoByLogin(context.Context, *UserInfoReq) (*UserInfoResp, error)
	// 创建 Group
	CreateGroup(context.Context, *CreateGroupReq) (*CreateGroupResp, error)
}

// UnimplementedYuqueServer should be embedded to have forward compatible implementations.
type UnimplementedYuqueServer struct {
}

func (UnimplementedYuqueServer) UserInfo(context.Context, *emptypb.Empty) (*UserInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserInfo not implemented")
}
func (UnimplementedYuqueServer) UserInfoByLogin(context.Context, *UserInfoReq) (*UserInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserInfoByLogin not implemented")
}
func (UnimplementedYuqueServer) CreateGroup(context.Context, *CreateGroupReq) (*CreateGroupResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroup not implemented")
}

// UnsafeYuqueServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to YuqueServer will
// result in compilation errors.
type UnsafeYuqueServer interface {
	mustEmbedUnimplementedYuqueServer()
}

func RegisterYuqueServer(s grpc.ServiceRegistrar, srv YuqueServer) {
	s.RegisterService(&Yuque_ServiceDesc, srv)
}

func _Yuque_UserInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(YuqueServer).UserInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.Yuque/UserInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(YuqueServer).UserInfo(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Yuque_UserInfoByLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(YuqueServer).UserInfoByLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.Yuque/UserInfoByLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(YuqueServer).UserInfoByLogin(ctx, req.(*UserInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Yuque_CreateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateGroupReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(YuqueServer).CreateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.Yuque/CreateGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(YuqueServer).CreateGroup(ctx, req.(*CreateGroupReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Yuque_ServiceDesc is the grpc.ServiceDesc for Yuque service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Yuque_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "yuque.v2.Yuque",
	HandlerType: (*YuqueServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UserInfo",
			Handler:    _Yuque_UserInfo_Handler,
		},
		{
			MethodName: "UserInfoByLogin",
			Handler:    _Yuque_UserInfoByLogin_Handler,
		},
		{
			MethodName: "CreateGroup",
			Handler:    _Yuque_CreateGroup_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/yuque/yuque.proto",
}

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	// user signin
	Signin(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error)
	Signin1(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error)
	// user resets password
	ResetPassword(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) Signin(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error) {
	out := new(UserInfoResp)
	err := c.cc.Invoke(ctx, "/yuque.v2.UserService/Signin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Signin1(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*UserInfoResp, error) {
	out := new(UserInfoResp)
	err := c.cc.Invoke(ctx, "/yuque.v2.UserService/Signin1", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) ResetPassword(ctx context.Context, in *UserInfoReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/yuque.v2.UserService/ResetPassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations should embed UnimplementedUserServiceServer
// for forward compatibility
type UserServiceServer interface {
	// user signin
	Signin(context.Context, *UserInfoReq) (*UserInfoResp, error)
	Signin1(context.Context, *UserInfoReq) (*UserInfoResp, error)
	// user resets password
	ResetPassword(context.Context, *UserInfoReq) (*emptypb.Empty, error)
}

// UnimplementedUserServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUserServiceServer struct {
}

func (UnimplementedUserServiceServer) Signin(context.Context, *UserInfoReq) (*UserInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Signin not implemented")
}
func (UnimplementedUserServiceServer) Signin1(context.Context, *UserInfoReq) (*UserInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Signin1 not implemented")
}
func (UnimplementedUserServiceServer) ResetPassword(context.Context, *UserInfoReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetPassword not implemented")
}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_Signin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Signin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.UserService/Signin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Signin(ctx, req.(*UserInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Signin1_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Signin1(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.UserService/Signin1",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Signin1(ctx, req.(*UserInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_ResetPassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).ResetPassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yuque.v2.UserService/ResetPassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).ResetPassword(ctx, req.(*UserInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "yuque.v2.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Signin",
			Handler:    _UserService_Signin_Handler,
		},
		{
			MethodName: "Signin1",
			Handler:    _UserService_Signin1_Handler,
		},
		{
			MethodName: "ResetPassword",
			Handler:    _UserService_ResetPassword_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/yuque/yuque.proto",
}