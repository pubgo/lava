package handler

import (
	"context"
	"fmt"
	hello2 "github.com/pubgo/lava/internal/example/services/protopb/proto/hello"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/plugins/db"
)

func NewTestAPIHandler() hello2.TestApiServer {
	return &testapiHandler{}
}

type testapiHandler struct {
	Db *db.Client `dix:""`
}

func (h *testapiHandler) Init() {

}

func (h *testapiHandler) VersionTestCustom(ctx context.Context, req *hello2.TestReq) (*hello2.TestApiOutput, error) {
	panic("implement me")
}

func (h *testapiHandler) Version1(ctx context.Context, value *structpb.Value) (*hello2.TestApiOutput1, error) {
	fmt.Printf("%#v\n", value.GetStructValue().AsMap())
	return &hello2.TestApiOutput1{
		Data: value,
	}, nil
}

func (h *testapiHandler) Version(ctx context.Context, in *hello2.TestReq) (out *hello2.TestApiOutput, err error) {
	var log = logger.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)

	if h.Db != nil {
		log.Info("dix db ok", zap.Any("err", h.Db.Get().Ping()))
		log.Info("dix config ok", zap.String("cfg", config.GetCfg().ConfigFileUsed()))
	}

	out = &hello2.TestApiOutput{
		Msg: in.Input,
	}
	time.Sleep(time.Millisecond * 10)
	return
}

func (h *testapiHandler) VersionTest(ctx context.Context, in *hello2.TestReq) (out *hello2.TestApiOutput, err error) {

	out = &hello2.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
