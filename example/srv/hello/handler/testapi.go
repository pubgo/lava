package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/example/protopb/hellopb"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
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

var ll = logging.Component("handler")

func NewTestAPIHandler(Db *orm.Client, Cron *scheduler.Scheduler, conns map[string]*grpcc.Client, L *logging.Logger) hellopb.TestApiServer {
	return &testApiHandler{
		Db:         Db,
		Cron:       Cron,
		TestApiSrv: hellopb.NewTestApiClient(conns["test-grpc"]),
		L:          L,
	}
}

type testApiHandler struct {
	Db         *orm.Client
	Cron       *scheduler.Scheduler
	TestApiSrv hellopb.TestApiClient
	L          *logging.Logger
}

func (h *testApiHandler) Close() {
	h.L.Info("close")
}

func (h *testApiHandler) Init() {
	defer xerror.RespExit()

	var db = h.Db

	xerror.Panic(db.AutoMigrate(&User{}))
	var user = User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	xerror.Panic(db.Create(&user).Error)

	logutil.ColorPretty(user)

	//buf := &bytes.Buffer{}
	//memviz.Map(buf, &user)
	//xerror.Panic(ioutil.WriteFile("example-tree-data", buf.Bytes(), 0644))

	h.Cron.Every("test grpc client", time.Second*5, func(name string) {
		defer xerror.RespExit()
		zap.L().Debug("客户端访问")
		var out, err1 = h.TestApiSrv.Version(context.Background(), &hellopb.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	})
}

func (h *testApiHandler) VersionTestCustom(ctx context.Context, req *hellopb.TestReq) (*hellopb.TestApiOutput, error) {
	panic("implement me")
}

func (h *testApiHandler) Version1(ctx context.Context, value *structpb.Value) (*hellopb.TestApiOutput1, error) {
	fmt.Printf("%#v\n", value.GetStructValue().AsMap())
	return &hellopb.TestApiOutput1{
		Data: value,
	}, nil
}

func (h *testApiHandler) Version(ctx context.Context, in *hellopb.TestReq) (out *hellopb.TestApiOutput, err error) {
	var log = logging.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)
	ll.S().Infof("Received Helloworld.Call request, name: %s", in.Input)

	var m = metric.GetMetric(ctx)
	m.Counter("test-counter").Inc(1)
	defer m.Timer("test-timer").Start().Stop()

	if h.Db != nil {
		var user User

		var db = h.Db

		xerror.Panic(db.WithContext(ctx).First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		xerror.Panic(db.Raw("select * from users limit 1").First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		xerror.Panic(db.Model(&User{}).Where("Age = ?", 18).First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		log.Sugar().Infow("config ok", "cfg", config.CfgPath)
	}

	out = &hellopb.TestApiOutput{
		Msg: in.Input,
	}

	if in.Input == "error" {
		return out, errors.New("error test")
	}

	return
}

func (h *testApiHandler) VersionTest(ctx context.Context, in *hellopb.TestReq) (out *hellopb.TestApiOutput, err error) {
	out = &hellopb.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
