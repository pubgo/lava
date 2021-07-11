package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/lug/registry"
	_ "github.com/pubgo/lug/registry/mdns"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"
)

var testApiSrv = hello.GetTestApiClient("test-grpc", func(service string) []grpc.DialOption {
	fmt.Println("service", service)
	return nil
})

func main() {
	registry.Init()

	_ = fx.Tick(func(ctx fx.Ctx) {
		xlog.Info("客户端访问")

		defer xerror.RespDebug()

		var cli, err = testApiSrv()
		xerror.Panic(err)

		var out, err1 = cli.Version(context.Background(), &hello.TestReq{Input: "input", Name: "hello"})
		xerror.Panic(err1)
		fmt.Printf("%#v \n", out)
	}, time.Second*5)
	select {}
}
