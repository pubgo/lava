// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/login/login.proto

package login

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc "github.com/pubgo/lava/clients/grpcc"
	service "github.com/pubgo/lava/service"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func InitLoginClient(srv string, opts ...func(cfg *grpcc.Cfg)) {
	grpcc.InitClient(srv, append(opts, grpcc.WithClientType((*LoginClient)(nil)))...)
}

func RegisterLogin(srv service.Service, impl LoginServer) {
	var desc service.Desc
	desc.Handler = impl
	desc.ServiceDesc = Login_ServiceDesc
	desc.GrpcClientFn = NewLoginClient

	desc.GrpcGatewayFn = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterUserServiceHandlerClient(ctx, mux, NewUserServiceClient(conn))
	}

	srv.RegisterService(desc)
}
