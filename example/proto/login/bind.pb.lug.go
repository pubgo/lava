// Code generated by protoc-gen-lug. DO NOT EDIT.
// source: example/proto/login/bind.proto

package login

import (
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	fb "github.com/pubgo/lug/builder/fiber"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/plugins/grpcc"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

var _ = strings.Trim
var _ = utils.UnsafeString
var _ fiber.Router = nil
var _ = gutil.MapFormByTag
var _ = fb.Cfg{}

func GetBindTelephoneClient(srv string, optFns ...func(service string) []grpc.DialOption) func() (BindTelephoneClient, error) {
	client := grpcc.GetClient(srv, optFns...)
	return func() (BindTelephoneClient, error) {
		c, err := client.Get()
		return &bindTelephoneClient{c}, xerror.WrapF(err, "srv: %s", srv)
	}
}

func init() {
	var mthList []xgen.GrpcRestHandler

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "Check",
		Method:       "POST",
		Path:         "/user/bind-telephone/check",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "BindVerify",
		Method:       "POST",
		Path:         "/user/bind-telephone/bind-verify",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "BindChange",
		Method:       "POST",
		Path:         "/user/bind-telephone/bind-change",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "AutomaticBind",
		Method:       "POST",
		Path:         "/user/bind-telephone/automatic-bind",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "BindPhoneParse",
		Method:       "POST",
		Path:         "/user/bind-telephone/bind-phone-parse",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.BindTelephone",
		Name:         "BindPhoneParseByOneClick",
		Method:       "POST",
		Path:         "/user/bind-telephone/bind-phone-parse-by-one-click",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	xgen.Add(reflect.ValueOf(RegisterBindTelephoneServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterBindTelephoneRestServer), nil)
	xgen.Add(reflect.ValueOf(RegisterBindTelephoneHandler), nil)
}

func RegisterBindTelephoneRestServer(app fiber.Router, server BindTelephoneServer) {
	if app == nil || server == nil {
		panic("app is nil or server is nil")
	}

	// restful
	app.Add("POST", "/user/bind-telephone/check", func(ctx *fiber.Ctx) error {
		var req = new(CheckRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.Check(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

	// restful
	app.Add("POST", "/user/bind-telephone/bind-verify", func(ctx *fiber.Ctx) error {
		var req = new(BindVerifyRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.BindVerify(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

	// restful
	app.Add("POST", "/user/bind-telephone/bind-change", func(ctx *fiber.Ctx) error {
		var req = new(BindChangeRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.BindChange(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

	// restful
	app.Add("POST", "/user/bind-telephone/automatic-bind", func(ctx *fiber.Ctx) error {
		var req = new(AutomaticBindRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.AutomaticBind(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

	// restful
	app.Add("POST", "/user/bind-telephone/bind-phone-parse", func(ctx *fiber.Ctx) error {
		var req = new(BindPhoneParseRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.BindPhoneParse(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

	// restful
	app.Add("POST", "/user/bind-telephone/bind-phone-parse-by-one-click", func(ctx *fiber.Ctx) error {
		var req = new(BindPhoneParseByOneClickRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.BindPhoneParseByOneClick(ctx.UserContext(), req)
		if err != nil {
			return err
		}

		return ctx.JSON(resp)
	})

}
