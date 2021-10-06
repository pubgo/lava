// Code generated by protoc-gen-lug. DO NOT EDIT.
// versions:
// - protoc-gen-lug v0.1.0
// - protoc         v3.17.3
// source: example/proto/hello/helloworld.proto

package hello

import (
	grpcc "github.com/pubgo/lug/plugins/grpcc"
	xgen "github.com/pubgo/lug/xgen"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetGreeterClient(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli GreeterClient)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli GreeterClient)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&greeterClient{c})
		return
	}
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
	xgen.Add(reflect.ValueOf(RegisterGreeterServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterGreeterHandlerServer), nil)
}