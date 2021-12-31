// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/login/merge.proto

package login

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

func GetMergeClient(srv string, opts ...func(cfg *grpcc.Cfg)) MergeClient {
	return &mergeClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "Telephone",
		Method:       "POST",
		Path:         "/user/merge/telephone",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TelephoneRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "TelephoneCheck",
		Method:       "POST",
		Path:         "/user/merge/telephone-check",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChat",
		Method:       "POST",
		Path:         "/user/merge/we-chat",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatCheck",
		Method:       "POST",
		Path:         "/user/merge/we-chat-check",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &WeChatUnMergeRequest{},
		Output:       &Reply{},
		Service:      "login.Merge",
		Name:         "WeChatUnMerge",
		Method:       "POST",
		Path:         "/user/merge/we-chat-un-merge",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterMergeServer, mthList)
	var registerMergeGrpcClient = func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
		return RegisterMergeHandlerClient(ctx, mux, NewMergeClient(conn))
	}
	xgen.Add(registerMergeGrpcClient, nil)
}
