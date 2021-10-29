// Code generated by protoc-gen-lava. DO NOT EDIT.
// versions:
// - protoc-gen-lava v0.1.0
// - protoc         v3.17.3
// source: proto/hello/api.proto

package hello

import (
	gin "github.com/gin-gonic/gin"
	fiber "github.com/pubgo/lava/builder/fiber"
	grpcc "github.com/pubgo/lava/clients/grpcc"
	binding "github.com/pubgo/lava/pkg/binding"
	xgen "github.com/pubgo/lava/xgen"
	byteutil "github.com/pubgo/x/byteutil"
	xerror "github.com/pubgo/xerror"
	grpc "google.golang.org/grpc"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

func GetTestApiClient(srv string, opts ...func(cfg *grpcc.Cfg)) TestApiClient {
	return &testApiClient{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApi",
		Name:         "Version",
		Method:       "GET",
		Path:         "/v1/version",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &structpb.Value{},
		Output:       &TestApiOutput1{},
		Service:      "hello.TestApi",
		Name:         "Version1",
		Method:       "POST",
		Path:         "/v1/version1",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApi",
		Name:         "VersionTest",
		Method:       "GET",
		Path:         "/v1/example/versiontest",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApi",
		Name:         "VersionTestCustom",
		Method:       "sql",
		Path:         "/sql",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterTestApiServer, mthList)
	xgen.Add(RegisterTestApiHandler, nil)
	xgen.Add(RegisterTestApiRestServer, nil)
	xgen.Add(RegisterTestApiGinServer, nil)
}
func RegisterTestApiRestServer(app fiber.Router, server TestApiServer) {
	xerror.Assert(app == nil || server == nil, "app or server is nil")
	app.Add("GET", "/v1/version", func(ctx *fiber.Ctx) error {
		var req = new(TestReq)
		data := make(map[string][]string)
		ctx.Context().QueryArgs().VisitAll(func(key []byte, val []byte) {
			k := byteutil.ToStr(key)
			v := byteutil.ToStr(val)
			data[k] = append(data[k], v)
		})
		xerror.Panic(binding.MapFormByTag(req, data, "json"))
		var resp, err = server.Version(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/v1/version1", func(ctx *fiber.Ctx) error {
		var req = new(structpb.Value)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.Version1(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("GET", "/v1/example/versiontest", func(ctx *fiber.Ctx) error {
		var req = new(TestReq)
		data := make(map[string][]string)
		ctx.Context().QueryArgs().VisitAll(func(key []byte, val []byte) {
			k := byteutil.ToStr(key)
			v := byteutil.ToStr(val)
			data[k] = append(data[k], v)
		})
		xerror.Panic(binding.MapFormByTag(req, data, "json"))
		var resp, err = server.VersionTest(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("SQL", "/sql", func(ctx *fiber.Ctx) error {
		var req = new(TestReq)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.VersionTestCustom(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
}
func RegisterTestApiGinServer(r gin.IRouter, server TestApiServer) {
	xerror.Assert(r == nil || server == nil, "router or server is nil")
	r.Handle("GET", "/v1/version", func(ctx *gin.Context) {
		var req = new(TestReq)
		xerror.Panic(binding.MapFormByTag(req, ctx.Request.URL.Query(), "json"))
		var resp, err = server.Version(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("POST", "/v1/version1", func(ctx *gin.Context) {
		var req = new(structpb.Value)
		xerror.Panic(ctx.ShouldBindJSON(req))
		var resp, err = server.Version1(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("GET", "/v1/example/versiontest", func(ctx *gin.Context) {
		var req = new(TestReq)
		xerror.Panic(binding.MapFormByTag(req, ctx.Request.URL.Query(), "json"))
		var resp, err = server.VersionTest(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("SQL", "/sql", func(ctx *gin.Context) {
		var req = new(TestReq)
		xerror.Panic(ctx.ShouldBindJSON(req))
		var resp, err = server.VersionTestCustom(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
}
func GetTestApiV2Client(srv string, opts ...func(cfg *grpcc.Cfg)) TestApiV2Client {
	return &testApiV2Client{grpcc.GetClient(srv, opts...)}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApiV2",
		Name:         "Version1",
		Method:       "POST",
		Path:         "/v2/example/version/{name}",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &TestReq{},
		Output:       &TestApiOutput{},
		Service:      "hello.TestApiV2",
		Name:         "VersionTest1",
		Method:       "POST",
		Path:         "/v2/example/versiontest",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(RegisterTestApiV2Server, mthList)
	xgen.Add(RegisterTestApiV2Handler, nil)
	xgen.Add(RegisterTestApiV2RestServer, nil)
	xgen.Add(RegisterTestApiV2GinServer, nil)
}
func RegisterTestApiV2RestServer(app fiber.Router, server TestApiV2Server) {
	xerror.Assert(app == nil || server == nil, "app or server is nil")
	app.Add("POST", "/v2/example/version/{name}", func(ctx *fiber.Ctx) error {
		var req = new(TestReq)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.Version1(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/v2/example/versiontest", func(ctx *fiber.Ctx) error {
		var req = new(TestReq)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.VersionTest1(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
}
func RegisterTestApiV2GinServer(r gin.IRouter, server TestApiV2Server) {
	xerror.Assert(r == nil || server == nil, "router or server is nil")
	r.Handle("POST", "/v2/example/version/{name}", func(ctx *gin.Context) {
		var req = new(TestReq)
		xerror.Panic(ctx.ShouldBindJSON(req))
		var resp, err = server.Version1(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
	r.Handle("POST", "/v2/example/versiontest", func(ctx *gin.Context) {
		var req = new(TestReq)
		xerror.Panic(ctx.ShouldBindJSON(req))
		var resp, err = server.VersionTest1(ctx, req)
		xerror.Panic(err)
		ctx.JSON(200, resp)
	})
}
