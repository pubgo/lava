package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	_ "github.com/pubgo/lava"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/example/protopb/hellopb"
	"github.com/pubgo/lava/logging"
)

var _srv *testApiHandler

func TestMain(t *testing.M) {
	defer xerror.RecoverAndExit()
	dix.Register(func(Db *orm.Client, Cron *scheduler.Scheduler, conns map[string]*grpcc.Client, L *logging.Logger) {
		_srv = &testApiHandler{
			Db:         Db,
			Cron:       Cron,
			TestApiSrv: hellopb.NewTestApiClient(conns["test-grpc"]),
			L:          L,
		}
	})

	dix.Invoke()

	_srv.Init()
	t.Run()
}

func TestInit(t *testing.T) {
	fmt.Println(_srv.Version(context.Background(), &hellopb.TestReq{
		Input: "hello",
	}))
}
