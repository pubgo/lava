package handler

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/logger"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pubgo/lava/clients/db"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/example/services/protopb/proto/hello"
	"github.com/pubgo/lava/middlewares/logRecord"
	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/middlewares/traceRecord"
	"github.com/pubgo/lava/plugins/scheduler"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(cfg *grpcc.Cfg) {
	cfg.Middlewares = append(cfg.Middlewares, requestID.Name)
	cfg.Middlewares = append(cfg.Middlewares, logRecord.Name)
	cfg.Middlewares = append(cfg.Middlewares, traceRecord.Name)
})

func NewTestAPIHandler() *testapiHandler {
	return &testapiHandler{}
}

type testapiHandler struct {
	Db   *xorm.Client         `dix:""`
	Cron *scheduler.Scheduler `dix:""`
}

func (h *testapiHandler) Init() {
	h.Cron.Every("test grpc client", time.Second*2, func(name string) {
		zap.L().Debug("客户端访问")
		var out, err1 = testApiSrv.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	})
}

func (h *testapiHandler) VersionTestCustom(ctx context.Context, req *hello.TestReq) (*hello.TestApiOutput, error) {
	panic("implement me")
}

func (h *testapiHandler) Version1(ctx context.Context, value *structpb.Value) (*hello.TestApiOutput1, error) {
	fmt.Printf("%#v\n", value.GetStructValue().AsMap())
	return &hello.TestApiOutput1{
		Data: value,
	}, nil
}

func (h *testapiHandler) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	var log = logger.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)

	if h.Db != nil {
		log.Info("dix db ok", logger.WithErr(h.Db.Ping())...)
		log.Info("dix config ok", zap.String("cfg", config.GetCfg().ConfigPath()))
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
