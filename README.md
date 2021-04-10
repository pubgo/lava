# [golug](https://www.notion.so/barrylog/golug-96142de3b0444b6c905886efac96eeb0)

1. protoc
    curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip
    go get github.com/golang/protobuf/protoc-gen-go@v1.3.2
    go install -v github.com/gogo/protobuf/protoc-gen-gofast
    go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
    go install -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
    go install -v github.com/vektra/mockery/cmd/mockery

1. 配置管理
2.

1. golug是一个高度抽象和集成的微服务框架
2. golug集成了config, log, command, plugin, grpc, http(fiber), task, broker等组件
3. golug使用方便, 统一入口
4. golug把http, grpc等server抽象成统一的entry, 统一使用习惯
5. golug统一运行入口, 让多个服务同时集成和运行
6. golug对grpc的protobuf进行定制处理, 让grpc server register更加方便
7. golug 抽象plugin, 跟随config的动态化加载,
8. golug config和watcher分离, config从本地文件加载, watcher可以从远程的任何的组件watch, 比如etcd


## 功能特性
1. 健康检查, pprof
2. 一切配置化, 通过plugin集成组建
3. 集成log, metric, tracing
4. 集成多种服务和server
5. 一站式开发
6. 环境自动检测, 本地项目生成
7. 支持rest, grpc等

## example

```go
package http_entry

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewHttpEntry("http")
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("entry http test"))

	ent.Use(func(ctx *fiber.Ctx) error {
		fmt.Println("ok")

		return ctx.Next()
	})

	ent.Group("/api", func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})

	return ent
}
```

```go
import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/ctl_entry"
	"github.com/pubgo/golug/example/grpc_entry"
	"github.com/pubgo/golug/example/http_entry"
	"github.com/pubgo/xerror"
)

func main() {
	xerror.Exit(golug.Init())
	xerror.Exit(golug.Run(
		http_entry.GetEntry(),
		ctl_entry.GetEntry(),
		grpc_entry.GetEntry(),
		grpc_entry.GetHttpEntry(),
	))
}
```


https://github.com/fullstorydev/grpcurl
https://github.com/bojand/ghz
