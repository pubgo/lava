package grpc_entry

import (
	"context"
	"fmt"
	"github.com/pubgo/lug/plugins/grpcc"
	"time"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/example/grpc_entry/handler"
	"github.com/pubgo/lug/example/proto/hello"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"
)

var name = "test-grpc"

var _ = hello.GetTestApiClient(name, func(service string) []grpc.DialOption {
	fmt.Println("service", service)
	return nil
})

func GetEntry() entry.Entry {
	ent := lug.NewGrpc(name)
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	ent.Init()
	ent.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//q.Q(info)
		return handler(ctx, req)
	})

	ent.AfterStart(func() {
		_ = fx.GoLoop(func(ctx context.Context) {
			xlog.Info("客户端访问")

			defer xerror.Resp(func(err xerror.XErr) {
				err.Debug()
				time.Sleep(time.Second * 5)
			})

			var conn, err = grpcc.NewDirect(":8080")
			xerror.Panic(err)
			//var cli, err = testApiSrv()
			//xerror.Panic(err)

			cli := hello.NewTestApiClient(conn)
			var out, err1 = cli.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
			xerror.Panic(err1)
			xlog.Infof("%s\n",out)

			time.Sleep(time.Second * 5)
		})
	})

	return ent
}
