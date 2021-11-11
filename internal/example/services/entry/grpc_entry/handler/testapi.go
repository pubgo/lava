package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/example/services/protopb/proto/hello"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/middlewares/logRecord"
	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/middlewares/traceRecord"
	"github.com/pubgo/lava/plugins/scheduler"
)

type User struct {
	gorm.Model
	ID           uint
	Name         string
	Email        *string
	Age          uint8
	Birthday     time.Time
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var testApiSrv = hello.GetTestApiClient("test-grpc", func(cfg *grpcc.Cfg) {
	cfg.Middlewares = append(cfg.Middlewares, requestID.Name)
	cfg.Middlewares = append(cfg.Middlewares, logRecord.Name)
	cfg.Middlewares = append(cfg.Middlewares, traceRecord.Name)
})

func NewTestAPIHandler() *testapiHandler {
	return &testapiHandler{}
}

type testapiHandler struct {
	Db   *orm.Client          `dix:""`
	Cron *scheduler.Scheduler `dix:""`
}

func (h *testapiHandler) Init() {
	xerror.Panic(h.Db.AutoMigrate(&User{}))
	var user = User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	xerror.Panic(h.Db.Create(&user).Error)
	q.Q(user)

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
		var user User
		xerror.Panic(h.Db.WithContext(ctx).First(&user).Error)
		log.Info("data", zap.Any("data", user))
		log.Info("dix db ok", logger.WithErr(h.Db.Raw("select 1").Error)...)
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
