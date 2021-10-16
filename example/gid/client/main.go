package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	_ "unsafe"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	_ "net/http/pprof"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/example/gid/protopb/proto/hello"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/grpcc"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/runenv"
	_ "github.com/pubgo/lava/tracing/jaeger"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(cfg *grpcc.Cfg) {
	fmt.Println("service", cfg)
})

func main() {
	go http.ListenAndServe(runenv.DebugAddr, nil)

	runenv.Project = "test-client"

	// 初始化配置
	xerror.Exit(config.Init())

	// 初始化注册中心
	xerror.Exit(plugin.Get(registry.Name).Init(nil))

	defer xerror.RespDebug()

	for range time.Tick(time.Second * 5) {
		zap.L().Debug("客户端访问")
		var out, err1 = testApiSrv.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	}
}
