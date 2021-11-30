// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/yuque/yuque.proto

package yuque_pb

import (
	gin "github.com/gin-gonic/gin"
	grpcc "github.com/pubgo/lava/clients/grpcc"
	binding "github.com/pubgo/lava/pkg/binding"
	xgen "github.com/pubgo/lava/xgen"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetYuqueClient(srv string, opts ...func(cfg *grpcc.Cfg)) YuqueClient {
	return &yuqueClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &emptypb.Empty{},
		Output:       &UserInfoResp{},
		Service:      "yuque.v2.Yuque",
		Name:         "UserInfo",
		Method:       "GET",
		Path:         "/user",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UserInfoReq{},
		Output:       &UserInfoResp{},
		Service:      "yuque.v2.Yuque",
		Name:         "UserInfoByLogin",
		Method:       "GET",
		Path:         "/users/{login}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &UserInfoReq{},
		Output:       &UserInfoResp{},
		Service:      "yuque.v2.Yuque",
		Name:         "UserInfoById",
		Method:       "GET",
		Path:         "/users/{id}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &CreateGroupReq{},
		Output:       &CreateGroupResp{},
		Service:      "yuque.v2.Yuque",
		Name:         "CreateGroup",
		Method:       "POST",
		Path:         "/groups",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterYuqueServer, mthList)
	xgen.Add(RegisterYuqueHandler, nil)
	xgen.Add(RegisterYuqueGinServer, nil)
}
func RegisterYuqueGinServer(r gin.IRouter, server YuqueServer) {
	xerror.Assert(r == nil || server == nil, "router or server is nil")
	r.Handle("GET", "/user", func(ctx *gin.Context) {
		var req = new(emptypb.Empty)
		xerror.Panic(binding.MapFormByTag(req, ctx.Request.URL.Query(), "json"))
		var resp, err = server.UserInfo(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("GET", "/users/{login}", func(ctx *gin.Context) {
		var req = new(UserInfoReq)
		xerror.Panic(binding.MapFormByTag(req, ctx.Request.URL.Query(), "json"))
		var resp, err = server.UserInfoByLogin(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("GET", "/users/{id}", func(ctx *gin.Context) {
		var req = new(UserInfoReq)
		xerror.Panic(binding.MapFormByTag(req, ctx.Request.URL.Query(), "json"))
		var resp, err = server.UserInfoById(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("POST", "/groups", func(ctx *gin.Context) {
		var req = new(CreateGroupReq)
		xerror.Panic(ctx.ShouldBindJSON(req))
		var resp, err = server.CreateGroup(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
}
