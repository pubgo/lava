package rest_entry

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/example/gid/proto/hello"
	"github.com/pubgo/lug/logger"
	db2 "github.com/pubgo/lug/plugins/db"
)

var _ hello.TestApiServer = (*Service)(nil)

type Service struct {
	Db  *db2.Client   `dix:""`
	Cfg config.Config `dix:""`
}

func (t *Service) VersionTestCustom(ctx context.Context, req *hello.TestReq) (*hello.TestApiOutput, error) {
	panic("implement me")
}

func (t *Service) Version1(ctx context.Context, req *structpb.Value) (*hello.TestApiOutput1, error) {
	panic("implement me")
}

func (t *Service) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	var log = logger.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)

	if t.Db != nil {
		log.Info("dix db ok", zap.Any("err", t.Db.Get().Ping()))
		log.Info("dix config ok", zap.String("cfg", t.Cfg.ConfigFileUsed()))
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
	rest.Provider(func(r rest.Router) {
		r.Use(func(ctx *fiber.Ctx) error {
			fmt.Println("ok")
			return ctx.Next()
		})

		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})
}
