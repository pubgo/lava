package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/lug/internal/debug"
	"github.com/pubgo/lug/registry"
	"github.com/pubgo/lug/registry/mdns"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"

	_ "net/http/pprof"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(service string) []grpc.DialOption {
	fmt.Println("service", service)
	return nil
})

func main() {
	go http.ListenAndServe(debug.Addr, nil)

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
