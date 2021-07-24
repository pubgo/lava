package rest_entry

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/db"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/example/proto/hello"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"
)

var _ rest.Service = (*Service)(nil)
var _ hello.TestApiServer = (*Service)(nil)

var logs = xlog.GetLogger("hello.handler")

type Service struct {
	Db  *db.Client    `dix:""`
	Cfg config.Config `dix:""`
}

func (t *Service) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	logs.Infof("Received Helloworld.Call request, name: %s", in.Input)

	if t.Db != nil {
		logs.Info("dix db ok", zap.Any("err", t.Db.Get().Ping()))
		logs.Info("dix config ok", zap.String("cfg", t.Cfg.ConfigFileUsed()))
	}

	out = &hello.TestApiOutput{
		Msg: in.Input,
	}
	time.Sleep(time.Millisecond * 10)
	return
}

func (t *Service) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {

	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}

func (t *Service) Router() func(r rest.Router) {
	return func(r rest.Router) {
		r.Use(func(ctx *fiber.Ctx) error {
			fmt.Println("ok")
			return ctx.Next()
		})

		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	}
}
