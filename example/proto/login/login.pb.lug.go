// Code generated by protoc-gen-lug. DO NOT EDIT.
// source: example/proto/login/login.proto

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
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = strings.Trim
var _ = utils.UnsafeString
var _ fiber.Router = nil
var _ = gutil.MapFormByTag
var _ = fb.Cfg{}
var _ = structpb.Value{}

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
		Service:      "login.Login",
		Name:         "Login",
		Method:       "POST",
		Path:         "/user/login/login",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	mthList = append(mthList, xgen.GrpcRestHandler{
		Service:      "login.Login",
		Name:         "Authenticate",
		Method:       "POST",
		Path:         "/user/login/authenticate",
		ClientStream: "False" == "True",
		ServerStream: "False" == "True",
		DefaultUrl:   "False" == "True",
	})

	xgen.Add(reflect.ValueOf(RegisterLoginServer), mthList)
	xgen.Add(reflect.ValueOf(RegisterLoginRestServer), nil)
	xgen.Add(reflect.ValueOf(RegisterLoginHandler), RegisterLoginServer)
}

func RegisterLoginRestServer(app fiber.Router, server LoginServer) {
	xerror.Assert(app == nil || server == nil, "app is nil or server is nil")

	// restful
	app.Add("POST", "/user/login/login", func(ctx *fiber.Ctx) error {
		var req = new(LoginRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.Login(ctx.UserContext(), req)
		if err != nil {
			return xerror.Wrap(err)
		}

		return xerror.Wrap(ctx.JSON(resp))
	})

	// restful
	app.Add("POST", "/user/login/authenticate", func(ctx *fiber.Ctx) error {
		var req = new(AuthenticateRequest)
		if err := ctx.BodyParser(req); err != nil {
			return xerror.Wrap(err)
		}

		var resp, err = server.Authenticate(ctx.UserContext(), req)
		if err != nil {
			return xerror.Wrap(err)
		}

		return xerror.Wrap(ctx.JSON(resp))
	})

}
