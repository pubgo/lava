package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/plugins/metric"
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

var testApiSrv = hello.GetTestApiClient("test-grpc")
var ll = logger.Component("handler")

func NewTestAPIHandler() *testapiHandler {
	return &testapiHandler{}
}

var _ = entry.AssertHandler(&testapiHandler{})

type testapiHandler struct {
	Db   *orm.Client          `dix:""`
	Cron *scheduler.Scheduler `dix:""`
}

func (h *testapiHandler) Init() {
	xerror.Panic(h.Db.Get().AutoMigrate(&User{}))
	var user = User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	xerror.Panic(h.Db.Get().Create(&user).Error)

	q.Q(user)

	//buf := &bytes.Buffer{}
	//memviz.Map(buf, &user)
	//xerror.Panic(ioutil.WriteFile("example-tree-data", buf.Bytes(), 0644))

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
	log.Infof("Received Helloworld.Call request, name: %s", in.Input)
	ll.Infof("Received Helloworld.Call request, name: %s", in.Input)

	var m = metric.GetWithCtx(ctx)
	m.Counter("test-counter").Inc(1)
	defer m.Timer("test-timer").Start().Stop()

	if h.Db != nil {
		var user User
		xerror.Panic(h.Db.Get().WithContext(ctx).First(&user).Error)
		log.Infow("data", "data", user)

		xerror.Panic(h.Db.Get().Raw("select * from users limit 1").First(&user).Error)
		log.Infow("data", "data", user)

		xerror.Panic(h.Db.Get().Model(&User{}).Where("Age = ?", 18).First(&user).Error)
		log.Infow("data", "data", user)

		log.Infow("dix config ok", "cfg", config.CfgPath)
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
