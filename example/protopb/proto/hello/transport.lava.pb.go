// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/hello/transport.proto

package hello

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

func InitTransportClient(srv string) {
	grpcc_builder.InitClient(srv, (*TransportClient)(nil), func(cc grpc.ClientConnInterface) interface{} { return NewTransportClient(cc) })
}

func RegisterTransport(srv service.Service, impl TransportServer) {
	var desc service.Desc
	desc.Handler = impl
	desc.ServiceDesc = Transport_ServiceDesc
	desc.GrpcClientFn = NewTransportClient

	desc.GrpcGatewayFn = func(mux *runtime.ServeMux) error {
		return RegisterTransportHandlerServer(context.Background(), mux, impl)
	}

	srv.RegisterService(desc)
}
