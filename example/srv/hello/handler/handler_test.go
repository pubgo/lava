package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/example/pkg/proto/hellopb"
	"github.com/pubgo/lava/logging"

	_ "github.com/pubgo/lava/plugins"
)

var _srv *testApiHandler

func TestMain(t *testing.M) {
	defer xerror.RecoverAndExit()
	dix.Register(func(Db *orm.Client, Cron *scheduler.Scheduler, conns map[string]*grpcc.Client, L *logging.Logger) {
		_srv = &testApiHandler{
			Db:         Db,
			Cron:       Cron,
			testApiSrv: hellopb.NewTestApiClient(conns["test-grpc"]),
			L:          L,
		}
	})

	_srv.Init()
	t.Run()
}

func TestInit(t *testing.T) {
	fmt.Println(_srv.Version(context.Background(), &hellopb.TestReq{
		Input: "hello",
	}))
}
