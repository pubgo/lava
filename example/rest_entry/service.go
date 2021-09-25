package rest_entry

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/db"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/lug/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ hello.TestApiServer = (*Service)(nil)

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) {
		logs = log.Named("hello.handler")
	})
}

type Service struct {
	Db  *db.Client    `dix:""`
	Cfg config.Config `dix:""`
}

func (t *Service) Version1(ctx context.Context, req *structpb.Value) (*hello.TestApiOutput1, error) {
	panic("implement me")
}

func (t *Service) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	logs.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)

	if t.Db != nil {
		logs.Info("dix db ok", zap.Any("err", t.Db.Get().Ping()))
		logs.Info("dix config ok", zap.String("cfg", t.Cfg.ConfigFileUsed()))
	}

	out = &hello.TestApiOutput{
		Msg: in.Input,
	}
	out.Reset()
	time.Sleep(time.Millisecond * 10)
	return
}

func (t *Service) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {

	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}

func init() {
	xerror.Exit(dix.Provider(func(r rest.Router) {
		r.Use(func(ctx *fiber.Ctx) error {
			fmt.Println("ok")
			return ctx.Next()
		})

		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	}))
}
