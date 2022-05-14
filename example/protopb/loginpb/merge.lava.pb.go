// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/login/merge.proto

package loginpb

import (
	context "context"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcc_builder "github.com/pubgo/lava/clients/grpcc/grpcc_builder"
	inject "github.com/pubgo/lava/inject"
	service "github.com/pubgo/lava/service"
	xgen "github.com/pubgo/lava/xgen"
	fx "go.uber.org/fx"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func InitMergeClient(addr string, alias ...string) {

	var name = ""
	if len(alias) > 0 {
		name = alias[0]
	}
	conn := grpcc_builder.NewClient(addr)

	inject.Register(fx.Provide(fx.Annotated{
		Target: func() MergeClient { return NewMergeClient(conn) },
		Name:   name,
	}))
}

func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "Telephone",
		Method:       "POST",
		Path:         "/login/merge/telephone",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "TelephoneCheck",
		Method:       "POST",
		Path:         "/login/merge/telephone-check",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChat",
		Method:       "POST",
		Path:         "/login/merge/we-chat",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatCheck",
		Method:       "POST",
		Path:         "/login/merge/we-chat-check",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatUnMergeRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatUnMerge",
		Method:       "POST",
		Path:         "/login/merge/we-chat-un-merge",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	xgen.Add(RegisterMergeServer, mthList)
}

func RegisterMerge(srv service.Service, impl MergeServer) {
	srv.RegService(service.Desc{
		Handler:     impl,
		ServiceDesc: Merge_ServiceDesc,
	})

	srv.RegGateway(func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error {
		return RegisterMergeHandlerClient(ctx, mux, NewMergeClient(cc))
	})

}