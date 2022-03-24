package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugins/metric"
	"github.com/pubgo/lava/plugins/scheduler"
	"github.com/pubgo/lava/service/service_type"
)

func init() {
	hello.InitTestApiClient("test-grpc", grpcc.WithDiscov())
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

var ll = logging.Component("handler")

func NewTestAPIHandler() *testapiHandler {
	return &testapiHandler{}
}

var _ service_type.Handler = (*testapiHandler)(nil)

type testapiHandler struct {
	Db         *orm.Client          `dix:""`
	Cron       *scheduler.Scheduler `dix:""`
	TestApiSrv hello.TestApiClient
	L          *logging.Logger `name:"testapiHandler"`
}

func (h *testapiHandler) Flags() typex.Flags { return nil }

func (h *testapiHandler) Router(r fiber.Router) {
}

func (h *testapiHandler) Init() func() {
	defer xerror.RespExit()

	var db = h.Db.Load()
	defer h.Db.Done()

	xerror.Panic(db.AutoMigrate(&User{}))
	var user = User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	xerror.Panic(db.Create(&user).Error)

	q.Q(user)

	//buf := &bytes.Buffer{}
	//memviz.Map(buf, &user)
	//xerror.Panic(ioutil.WriteFile("example-tree-data", buf.Bytes(), 0644))

	h.Cron.Every("test grpc client", time.Second*5, func(name string) {
		zap.L().Debug("客户端访问")
		var out, err1 = h.TestApiSrv.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	})

	return func() {
		h.L.Info("close")
	}
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
	var log = logging.GetLog(ctx)
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

func (h *testapiHandler) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
