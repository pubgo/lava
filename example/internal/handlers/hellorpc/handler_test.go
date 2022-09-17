package hellorpc

import (
	"context"
	"fmt"
	"github.com/pubgo/dix/di"
	"testing"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/example/gen/proto/hellopb"
	"github.com/pubgo/lava/logging"
)

var _srv *testApiHandler

func TestMain(t *testing.M) {
	defer xerror.RecoverAndExit()
	di.Provide(func(Db *orm.Client, Cron *scheduler.Scheduler, conns map[string]*grpcc.Client, L *logging.Logger) {
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
