// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/hello/helloworld.proto

package hello

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc "github.com/pubgo/lava/clients/grpcc"
	xgen "github.com/pubgo/lava/xgen"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetGreeterClient(srv string, opts ...func(cfg *grpcc.Cfg)) GreeterClient {
	return &greeterClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &HelloRequest{},
		Output:       &HelloReply{},
		Service:      "hello.Greeter",
		Name:         "SayHello",
		Method:       "GET",
		Path:         "/say/{name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterGreeterServer, mthList)
	var registerGreeterGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterGreeterHandlerClient(ctx, mux, NewGreeterClient(conn))
	}
	xgen.Add(registerGreeterGrpcClient, nil)
}
