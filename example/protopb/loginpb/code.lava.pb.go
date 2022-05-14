// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.19.4
// source: proto/login/code.proto

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

func InitCodeClient(addr string, alias ...string) {

	var name = ""
	if len(alias) > 0 {
		name = alias[0]
	}
	conn := grpcc_builder.NewClient(addr)

	inject.Register(fx.Provide(fx.Annotated{
		Target: func() CodeClient { return NewCodeClient(conn) },
		Name:   name,
	}))
}

func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &SendCodeRequest{},
		Output:       &SendCodeResponse{},
		Service:      "login.Code",
		Name:         "SendCode",
		Method:       "POST",
		Path:         "/login/code/send-code",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &VerifyRequest{},
		Output:       &VerifyResponse{},
		Service:      "login.Code",
		Name:         "Verify",
		Method:       "POST",
		Path:         "/login/code/verify",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &IsCheckImageCodeRequest{},
		Output:       &IsCheckImageCodeResponse{},
		Service:      "login.Code",
		Name:         "IsCheckImageCode",
		Method:       "POST",
		Path:         "/login/code/is-check-image-code",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &VerifyImageCodeRequest{},
		Output:       &VerifyImageCodeResponse{},
		Service:      "login.Code",
		Name:         "VerifyImageCode",
		Method:       "POST",
		Path:         "/login/code/verify-image-code",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &GetSendStatusRequest{},
		Output:       &GetSendStatusResponse{},
		Service:      "login.Code",
		Name:         "GetSendStatus",
		Method:       "POST",
		Path:         "/login/code/get-send-status",
		DefaultUrl:   true,
		ClientStream: false,
		ServerStream: false,
	})

	xgen.Add(RegisterCodeServer, mthList)
}

func RegisterCode(srv service.Service, impl CodeServer) {
	srv.RegService(service.Desc{
		Handler:     impl,
		ServiceDesc: Code_ServiceDesc,
	})

	srv.RegGateway(func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error {
		return RegisterCodeHandlerClient(ctx, mux, NewCodeClient(cc))
	})

}