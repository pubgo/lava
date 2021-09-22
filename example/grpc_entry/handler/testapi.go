package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/db"
	"github.com/pubgo/lug/example/proto/hello"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewTestAPIHandler() hello.TestApiServer {
	return &testapiHandler{}
}

type testapiHandler struct {
	Db  *db.Client    `dix:""`
	Cfg config.Config `dix:""`
}

func (h *testapiHandler) Version1(ctx context.Context, value *structpb.Value) (*hello.TestApiOutput1, error) {
	fmt.Printf("%#v\n", value.GetStructValue().AsMap())
	return &hello.TestApiOutput1{
		Data: value,
	}, nil
}

func (h *testapiHandler) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	log.Infof("Received Helloworld.Call request, name: %s", in.Input)

	if h.Db != nil {
		log.Info("dix db ok", zap.Any("err", h.Db.Get().Ping()))
		log.Info("dix config ok", zap.String("cfg", h.Cfg.ConfigFileUsed()))
	}

	out = &hello.TestApiOutput{
		Msg: in.Input,
	}
	time.Sleep(time.Millisecond * 10)
	return
}

func (h *testapiHandler) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {

	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
