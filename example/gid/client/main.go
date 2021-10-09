package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	_ "unsafe"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	_ "net/http/pprof"

	"github.com/pubgo/lug/example/gid/proto/hello"
	"github.com/pubgo/lug/plugins/grpcc"
	"github.com/pubgo/lug/plugins/registry"
	"github.com/pubgo/lug/plugins/registry/mdns"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/tracing"
	_ "github.com/pubgo/lug/tracing/jaeger"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(cfg *grpcc.Cfg) {
	cfg.Middlewares = append(cfg.Middlewares, tracing.Middleware)

	fmt.Println("service", cfg)
})

func main() {
	go http.ListenAndServe(runenv.DebugAddr, nil)

	runenv.Project = "test-client"

	var cfg = tracing.GetDefaultCfg()
	xerror.Exit(cfg.Build())

	xerror.Exit(registry.Init(mdns.Name, nil))

	_ = fx.Tick(func(ctx fx.Ctx) {
		zap.L().Debug("客户端访问")

		defer xerror.RespDebug()

		xerror.Panic(testApiSrv(func(cli hello.TestApiClient) {
			var out, err1 = cli.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
			xerror.Panic(err1)
			fmt.Printf("%#v \n", out)
		}))

	}, time.Second*5)
	select {}
}