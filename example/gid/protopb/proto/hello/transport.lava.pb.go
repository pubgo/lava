// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/hello/transport.proto

package hello

import (
	fiber "github.com/pubgo/lava/pkg/builder/fiber"
	grpcc "github.com/pubgo/lava/plugins/grpcc"
	xgen "github.com/pubgo/lava/xgen"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetTransportClient(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli TransportClient)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli TransportClient)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&transportClient{c})
		return
	}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &Message{},
		Output:       &Message{},
		Service:      "hello.Transport",
		Name:         "TestStream",
		Method:       "POST",
		Path:         "/hello/transport/test-stream",
		DefaultUrl:   true,
		ClientStream: true,
		ServerStream: true,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &Message{},
		Output:       &Message{},
		Service:      "hello.Transport",
		Name:         "TestStream1",
		Method:       "POST",
		Path:         "/hello/transport/test-stream1",
		DefaultUrl:   true,
		ClientStream: true,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &Message{},
		Output:       &Message{},
		Service:      "hello.Transport",
		Name:         "TestStream2",
		Method:       "POST",
		Path:         "/hello/transport/test-stream2",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: true,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &Message{},
		Output:       &Message{},
		Service:      "hello.Transport",
		Name:         "TestStream3",
		Method:       "POST",
		Path:         "/hello/transport/test-stream3",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(reflect.ValueOf(RegisterTransportServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterTransportRestServer), mthList)
}
func RegisterTransportRestServer(app fiber.Router, server TransportServer) {
	xerror.Assert(app == nil || server == nil, "app or server is nil")
	app.Add("POST", "/hello/transport/test-stream3", func(ctx *fiber.Ctx) error {
		var req = new(Message)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.TestStream3(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
}
