// Code generated by protoc-gen-lug. DO NOT EDIT.
// versions:
// - protoc-gen-lug v0.1.0
// - protoc         v3.17.3
// source: proto/login/login.proto

package login

import (
	fiber "github.com/pubgo/lug/pkg/builder/fiber"
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

func GetLoginClient(srv string, opts ...func(cfg *grpcc.Cfg)) func(func(cli LoginClient)) error {
	client := grpcc.GetClient(srv, opts...)
	return func(fn func(cli LoginClient)) (err error) {
		defer xerror.RespErr(&err)

		c, err := client.Get()
		if err != nil {
			return xerror.WrapF(err, "srv: %s", srv)
		}

		fn(&loginClient{c})
		return
	}
}
func init() {
	var mthList []xgen.GrpcRestHandler
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &LoginRequest{},
		Output:       &LoginResponse{},
		Service:      "login.Login",
		Name:         "Login",
		Method:       "POST",
		Path:         "/user/login/login",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	mthList = append(mthList, xgen.GrpcRestHandler{
		Input:        &AuthenticateRequest{},
		Output:       &AuthenticateResponse{},
		Service:      "login.Login",
		Name:         "Authenticate",
		Method:       "POST",
		Path:         "/user/login/authenticate",
		DefaultUrl:   false,
		ClientStream: false,
		ServerStream: false,
	})
	xgen.Add(reflect.ValueOf(RegisterLoginServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterLoginHandlerServer), nil)
}
func RegisterLoginRestServer(app fiber.Router, server LoginServer) {
	xerror.Assert(app == nil || server == nil, "app or server is nil")
	app.Add("POST", "/user/login/login", func(ctx *fiber.Ctx) error {
		var req = new(LoginRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.Login(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
	app.Add("POST", "/user/login/authenticate", func(ctx *fiber.Ctx) error {
		var req = new(AuthenticateRequest)
		xerror.Panic(ctx.BodyParser(req))
		var resp, err = server.Authenticate(ctx.UserContext(), req)
		xerror.Panic(err)
		return xerror.Wrap(ctx.JSON(resp))
	})
}
