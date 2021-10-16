// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/login/merge.proto

package login

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
	xgen.Add(reflect.ValueOf(RegisterMergeServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterMergeRestServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterMergeHandlerServer), nil)
}
func RegisterMergeRestServer(app fiber.Router, server MergeServer) {
	xerror.Assert(app == nil || server == nil, "app or server is nil")
	app.Add("POST", "/user/merge/telephone", func(ctx *fiber.Ctx) error {
		var req = new(TelephoneRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.Telephone(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/user/merge/telephone-check", func(ctx *fiber.Ctx) error {
		var req = new(TelephoneRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.TelephoneCheck(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/user/merge/we-chat", func(ctx *fiber.Ctx) error {
		var req = new(WeChatRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.WeChat(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/user/merge/we-chat-check", func(ctx *fiber.Ctx) error {
		var req = new(WeChatRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.WeChatCheck(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/user/merge/we-chat-un-merge", func(ctx *fiber.Ctx) error {
		var req = new(WeChatUnMergeRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.WeChatUnMerge(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
}
