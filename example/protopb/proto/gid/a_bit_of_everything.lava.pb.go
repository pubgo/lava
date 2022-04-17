// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/gid/a_bit_of_everything.proto

package gid

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc_builder "github.com/pubgo/lava/clients/grpcc/grpcc_builder"
	service "github.com/pubgo/lava/service"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func InitLoginServiceClient(srv string) {
	grpcc_builder.InitClient(srv, (*LoginServiceClient)(nil), func(cc grpc.ClientConnInterface) interface{} { return NewLoginServiceClient(cc) })
}

func RegisterLoginService(srv service.Service, impl LoginServiceServer) {
	var desc service.Desc
	desc.Handler = impl
	desc.ServiceDesc = LoginService_ServiceDesc
	desc.GrpcClientFn = NewLoginServiceClient

	desc.GrpcGatewayFn = func(mux *runtime.ServeMux) error {
		return RegisterLoginServiceHandlerServer(context.Background(), mux, impl)
	}

	srv.RegisterService(desc)
}
