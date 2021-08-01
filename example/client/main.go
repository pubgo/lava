package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/lug/internal/debug"
	"github.com/pubgo/lug/plugins/grpcc"
	"github.com/pubgo/lug/registry"
	"github.com/pubgo/lug/registry/mdns"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/tracing"
	_ "github.com/pubgo/lug/tracing/jaeger"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	_ "net/http/pprof"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(cfg *grpcc.Cfg) {
	cfg.Middlewares = append(cfg.Middlewares, tracing.Middleware())
	fmt.Println("service", cfg)
})

func main() {
	go http.ListenAndServe(debug.Addr, nil)

	runenv.Project = "test-client"

	var cfg = tracing.GetDefaultCfg()
	xerror.Exit(cfg.Build())

	xerror.Exit(registry.Init(mdns.Name, nil))

	_ = fx.Tick(func(ctx fx.Ctx) {
		xlog.Debug("客户端访问")

		defer xerror.RespDebug()

		var cli, err = testApiSrv()
		xerror.Panic(err)

		var out, err1 = cli.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	}, time.Second*5)
	select {}
}
