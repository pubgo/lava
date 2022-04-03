package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	logging2 "github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugins/scheduler"
	"github.com/pubgo/lava/service"
)

func init() {
	hello.InitTestApiClient("test-grpc", grpcc.WithDiscov())
	//hello.InitTestApiClient("localhost:8080", grpcc.WithDirect())
}

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

var ll = logging2.Component("handler")

func NewTestAPIHandler() *testApiHandler {
	return &testApiHandler{}
}

var _ service.Handler = (*testApiHandler)(nil)

type testApiHandler struct {
	Db         *orm.Client
	Cron       *scheduler.Scheduler
	TestApiSrv hello.TestApiClient
	L          *logging2.Logger `name:"testApiHandler"`
}

func (h *testApiHandler) Flags() typex.Flags { return nil }

func (h *testApiHandler) Router(r fiber.Router) {
}

func (h *testApiHandler) Close() {
	h.L.Info("close")
}

func (h *testApiHandler) Init() {
	defer xerror.RespExit()

	var db = h.Db.Load()
	defer h.Db.Done()

	xerror.Panic(db.AutoMigrate(&User{}))
	var user = User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	xerror.Panic(db.Create(&user).Error)

	logutil.ColorPretty(user)

	//buf := &bytes.Buffer{}
	//memviz.Map(buf, &user)
	//xerror.Panic(ioutil.WriteFile("example-tree-data", buf.Bytes(), 0644))

	h.Cron.Every("test grpc client", time.Second*5, func(name string) {
		zap.L().Debug("客户端访问")
		var out, err1 = h.TestApiSrv.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	})
}

func (h *testApiHandler) VersionTestCustom(ctx context.Context, req *hello.TestReq) (*hello.TestApiOutput, error) {
	panic("implement me")
}

func (h *testApiHandler) Version1(ctx context.Context, value *structpb.Value) (*hello.TestApiOutput1, error) {
	fmt.Printf("%#v\n", value.GetStructValue().AsMap())
	return &hello.TestApiOutput1{
		Data: value,
	}, nil
}

func (h *testApiHandler) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	var log = logging2.GetLog(ctx)
	log.Sugar().Infof("Received Helloworld.Call request, name: %s", in.Input)
	ll.S().Infof("Received Helloworld.Call request, name: %s", in.Input)

	var m = metric.GetMetric(ctx)
	m.Counter("test-counter").Inc(1)
	defer m.Timer("test-timer").Start().Stop()

	if h.Db != nil {
		var user User

		var db = h.Db.Load()
		defer h.Db.Done()

		xerror.Panic(db.WithContext(ctx).First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		xerror.Panic(db.Raw("select * from users limit 1").First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		xerror.Panic(db.Model(&User{}).Where("Age = ?", 18).First(&user).Error)
		log.Sugar().Infow("data", "data", user)

		log.Sugar().Infow("dix config ok", "cfg", config.CfgPath)
	}

	out = &hello.TestApiOutput{
		Msg: in.Input,
	}

	if in.Input == "error" {
		return out, errors.New("error test")
	}

	return
}

func (h *testApiHandler) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
