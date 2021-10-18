package rest_entry

import (
	"context"
	"fmt"
	hello2 "github.com/pubgo/lava/internal/example/services/protopb/proto/hello"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry/restEntry"
	"github.com/pubgo/lava/logger"
	db2 "github.com/pubgo/lava/plugins/db"
)

var _ hello2.TestApiServer = (*Service)(nil)

type Service struct {
	Db  *db2.Client   `dix:""`
	Cfg config.Config `dix:""`
}

func (t *Service) VersionTestCustom(ctx context.Context, req *hello2.TestReq) (*hello2.TestApiOutput, error) {
	panic("implement me")
}

func (t *Service) Version1(ctx context.Context, req *structpb.Value) (*hello2.TestApiOutput1, error) {
	panic("implement me")
}

func (t *Service) Version(ctx context.Context, in *hello2.TestReq) (out *hello2.TestApiOutput, err error) {
	var log = logger.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)

	if t.Db != nil {
		log.Info("dix db ok", zap.Any("err", t.Db.Get().Ping()))
		log.Info("dix config ok", zap.String("cfg", t.Cfg.ConfigFileUsed()))
	}

	out = &hello2.TestApiOutput{
		Msg: in.Input,
	}
	out.Reset()
	time.Sleep(time.Millisecond * 10)
	return
}

func (t *Service) VersionTest(ctx context.Context, in *hello2.TestReq) (out *hello2.TestApiOutput, err error) {

	out = &hello2.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}

func init() {
	restEntry.Provider(func(r restEntry.Router) {
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
