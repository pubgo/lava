# zrpc 使用说明

`zrpc` 是 Lava 中新增的 **protobuf RPC over NATS request-reply** 能力，支持 unary 与 streaming。

它的定位不是替代 gRPC，而是补齐一条更轻量的内部 RPC 通道：

- 消息体使用 protobuf 二进制编码
- 传输层使用 NATS request-reply / inbox streaming
- 支持 unary、server streaming、client streaming、bidi streaming
- 服务端/客户端都可复用 `lava.Middleware`

## 组件位置

| 路径                          | 说明                                                            |
| ----------------------------- | --------------------------------------------------------------- |
| `proto/zrpc/v1/options.proto` | zrpc method 扩展定义（subject / queue / timeout）               |
| `pkg/zrpc/`                   | zrpc Go runtime（client / server / error / request / response） |
| `clients/zrpcc/`              | Lava 风格 zrpc 客户端封装                                       |
| `servers/zrpcs/`              | Lava 风格 zrpc 服务宿主                                         |
| `tools/protoc-gen-zrpc-go/`   | 本地 protoc 插件入口                                            |
| `internal/zrpcgen/`           | zrpc Go 代码生成逻辑                                            |
| `internal/examples/zrpcdemo/` | 最小 demo 服务与测试                                            |

## 运行关系

```mermaid
flowchart LR
    Proto[proto/zrpcdemo/v1/*.proto] --> Gen[protobuild gen]
    Gen --> Bindings[pkg/proto/*_zrpc.pb.go]
    Bindings --> Zrpcs[servers/zrpcs]
    Bindings --> Zrpcc[clients/zrpcc]
    Zrpcs --> Runtime[pkg/zrpc]
    Zrpcc --> Runtime
    Runtime --> NATS[NATS request-reply]
```

## Proto 写法

在 `.proto` 中引入 `zrpc/v1/options.proto`，即可为某个 rpc 显式配置 zrpc 绑定信息。

```proto
syntax = "proto3";

package zrpcdemo.v1;

option go_package = "github.com/pubgo/lava/v2/pkg/proto/zrpcdemov1;zrpcdemov1";

import "zrpc/v1/options.proto";

service EchoService {
  rpc Echo(EchoRequest) returns (EchoResponse) {
    option (zrpc.v1.method) = {
      subject: "svc.zrpcdemo.EchoService/Echo"
      queue: "zrpcdemo.echo"
      timeout_ms: 1500
    };
  }

  rpc Reverse(EchoRequest) returns (EchoResponse);
}
```

如果不显式指定，当前默认规则如下：

- `subject`: `svc.{package}.{Service}/{Method}`
- `queue`: `{package}.{service}`（小写）
- `timeout_ms`: `3000`

## 生成代码

仓库根目录执行：

```bash
task proto:gen
```

该任务会先安装本地插件：

- `go install ./tools/protoc-gen-zrpc-go`

再通过 `protobuild gen` 生成：

- `pkg/proto/.../*.pb.go`
- `pkg/proto/.../*_grpc.pb.go`
- `pkg/proto/.../*_zrpc.pb.go`

## 生成产物会提供什么

以 `EchoService` 为例，生成代码会提供：

- `EchoServiceZrpcServer`
- `RegisterEchoServiceZrpcRoutes(...)`
- `RegisterEchoServiceZrpcServer(...)`
- `EchoServiceZrpcClient`
- `NewEchoServiceZrpcClient(...)`

以及每个方法的绑定常量：

- `EchoService_EchoSubject`
- `EchoService_EchoQueue`
- `EchoService_EchoTimeout`

## 服务端接入方式

### 方式 1：直接使用生成的 zrpc server 注册

```go
nc, _ := nats.Connect(nats.DefaultURL)
defer nc.Close()

srv, err := zrpcdemov1.RegisterEchoServiceZrpcServer(nc, echoService{}, "")
if err != nil {
    panic(err)
}
defer srv.Close()
```

### 方式 2：使用 Lava 的 `servers/zrpcs`

这是当前更推荐的框架接入方式：

```go
svc := zrpcs.New(zrpcs.Params{
    Log:    logger,
    Metric: tally.NoopScope,
    Conf:   &zrpcs.Config{URL: "nats://127.0.0.1:4222"},
    Registers: []zrpcs.RegisterFunc{
        func(srv *zrpc.Server) error {
            return zrpcdemov1.RegisterEchoServiceZrpcRoutes(srv, echoService{}, "")
        },
    },
})

if err := svc.Serve(signals.Context()); err != nil {
    panic(err)
}
```

`zrpcs` 默认会挂上：

- `serviceinfo`
- `metric`
- `accesslog`
- `recovery`

## 客户端接入方式

### 方式 1：直接使用生成的 typed client

```go
nc, _ := nats.Connect(nats.DefaultURL)
defer nc.Close()

cli := zrpcdemov1.NewEchoServiceZrpcClient(nc)
resp, err := cli.Echo(ctx, &zrpcdemov1.EchoRequest{Message: "hello"})
```

### 方式 2：使用 Lava 的 `clients/zrpcc`

```go
baseCli := zrpcc.New(&zrpcc.Config{
    URL:     "nats://127.0.0.1:4222",
    Timeout: time.Second,
}, zrpcc.Params{
    Log:    logger,
    Metric: tally.NoopScope,
})
defer baseCli.Close()

nc, err := baseCli.Conn()
if err != nil {
    panic(err)
}

typedCli := zrpcdemov1.NewEchoServiceZrpcClient(nc)
resp, err := typedCli.Reverse(ctx, &zrpcdemov1.EchoRequest{Message: "hello"})
```

### Streaming 调用示例

```go
// server streaming：发送一个请求，持续 Recv
ss, err := cli.EchoStream(ctx, &zrpcdemov1.EchoRequest{Message: "Hi"})
if err != nil {
    return err
}
for {
    msg, err := ss.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    fmt.Println(msg.GetMessage())
}

// client streaming：持续 Send，最后 CloseAndRecv
cs, err := cli.Collect(ctx)
if err != nil {
    return err
}
_ = cs.Send(&zrpcdemov1.EchoRequest{Message: "a"})
resp, err := cs.CloseAndRecv()
```

`zrpcc` 同样默认挂上：

- `serviceinfo`
- `metric`
- `accesslog`
- `recovery`

## Demo 运行

### 1）先启动 NATS

本地需先有 `nats-server`，默认地址：

- `nats://127.0.0.1:4222`

### 2）启动 demo 服务

```bash
go run ./internal/examples/zrpcdemo
```

可选地指定：

```bash
NATS_URL=nats://127.0.0.1:4222 go run ./internal/examples/zrpcdemo
```

### 3）运行 demo 测试

```bash
go test ./internal/examples/zrpcdemo
```

如果本地没有 `nats-server`，该测试会自动 `Skip`。

## 当前能力边界

当前 `zrpc` 已支持：

- protobuf unary request / response
- server / client / bidi streaming
- method 级 subject / queue / timeout 配置
- 统一 Go runtime
- `zrpcs` / `zrpcc` 框架接入
- generated typed client / server
- `lava.Middleware` 复用（accesslog / metric / recovery）

当前仍未覆盖：

- HTTP -> zrpc bridge
- gRPC -> zrpc proxy
- 更高层的 `lava.ZrpcRouter` 抽象

## Streaming 协议

streaming 基于 NATS inbox，帧类型通过 header `Zrpc-Stream-Frame` 区分：

| 帧 | 方向 | 说明 |
| --- | --- | --- |
| `open` | client → server | 携带 `Zrpc-Stream-Req-Subject`（请求侧 inbox） |
| `ack` | server → client | 握手成功 |
| `data` | 双向 | protobuf 消息体 |
| `end` | 双向 | 半关闭（CloseSend） |
| `error` | 双向 | `Zrpc-Status-Code` / `Zrpc-Status-Message` |

超时通过 header `Timeout` 传递（如 `3s`），服务端用于整个 stream session；客户端在 `OpenStream` 时也会为 stream 设置相同 deadline。

## 日志与可观测性

`pkg/zrpc` **不直接打业务日志**，而是通过 `lava.Middleware` 记录每次 RPC：

| 层级 | 日志来源 | 典型字段 |
| --- | --- | --- |
| `pkg/zrpc` | 无直接日志 | 由 middleware 包装 request/response |
| `servers/zrpcs` | 启动/停止 | `url`、`registers` |
| `clients/zrpcc` | 连接/关闭 | `url` |
| middleware | accesslog | `request_id`、`operation`（subject）、`service`、`latency`、`error` |

### 默认中间件

`zrpcs` / `zrpcc` 默认挂载：

- `serviceinfo`：注入客户端/服务信息
- `metric`：请求耗时与结果指标
- `accesslog`：成功 debug、失败 info，含 subject 与 request_id
- `recovery`：panic 转错误

### 日志示例

服务启动：

```text
level=info service=zrpc-server url=nats://127.0.0.1:4222 registers=1 msg="zrpc server started"
```

客户端连接：

```text
level=info service=zrpcc url=nats://127.0.0.1:4222 msg="zrpc client connected"
```

RPC 访问（由 accesslog 输出，字段因配置而异）：

```text
level=debug request_id=... operation=svc.zrpcdemo.EchoService/Echo service=zrpcdemo.EchoService latency=12 client=true
```

### 排障建议

1. 确认 NATS 可达：`clients/zrpcc.Healthy` 或 `nats-server` 日志
2. 用 `request_id` 串联客户端 accesslog 与服务端 accesslog
3. streaming 卡住时检查是否收到 `end` / `error` 帧，以及 `Timeout` header 是否过短
4. subject / queue 不匹配时客户端通常表现为超时，而非明确错误码

## 客户端生命周期

`zrpcc` 懒连接 NATS，`Close()` 后会标记为 closed，**不会**在后续调用中自动重连。关闭后应新建客户端实例。

```go
cli := zrpcc.New(cfg, params)
defer cli.Close()

// Close 之后 CallUnary / OpenStream 返回 "client is closed"
```

## CI 与测试

GitHub Actions `lint-test.yml` 在 `v2` 分支会：

- 启动 `nats:2.10` service container
- 运行全量 `go test`（含 `pkg/zrpc`、`clients/zrpcc`、`internal/examples/zrpcdemo`）

本地等价命令：

```bash
task test
go test ./pkg/zrpc/... ./clients/zrpcc/... ./internal/examples/zrpcdemo/...
```

## 推荐开发顺序

1. 先写 proto（必要时加 `(zrpc.v1.method)`）
2. 执行 `task proto:gen`
3. 服务端实现 `*ZrpcServer` 接口
4. 用 `Register...ZrpcRoutes(...)` 接到 `zrpcs`
5. 用 `zrpcc` 或生成的 typed client 验证调用