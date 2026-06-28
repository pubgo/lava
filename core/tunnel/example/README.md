# core/tunnel 示例集

每个子目录是一个独立可运行的 demo，单进程内跑通完整链路，方便拷贝改造。

| 目录 | 主题 | 演示的 API |
| --- | --- | --- |
| [`example/main.go`](./main.go) | 基础：gateway / agent / backend 多模式 | `tunnelgateway.New` / `tunnelagent.New` |
| [`quickstart/`](./quickstart) | 最简上手（推荐先看） | `tunnelagent.Standalone` / `tunnel.GetService` / `tunnel.ServiceURL` |
| [`auth/`](./auth) | 代理鉴权（401 vs 200） | `tunnel.ConfigureGatewayAuth` / `GetService(token)` |
| [`grpc/`](./grpc) | gRPC over tunnel | `tunnel.GRPCContextDialer` / `GRPCDialOptions` |
| [`tls/`](./tls) | TLS 加密隧道 | `GatewayConfig.TLS` / `AgentConfig.TLS`（含 `MinVersion`） |

## 运行

```bash
go run ./core/tunnel/example/quickstart
go run ./core/tunnel/example/auth
go run ./core/tunnel/example/grpc
go run ./core/tunnel/example/tls

# 基础多模式示例（默认单进程跑全部）
go run ./core/tunnel/example            # = main.go, mode=all
go run ./core/tunnel/example -mode=gateway
go run ./core/tunnel/example -transport=quic
```

各 demo 使用独立端口段，互不冲突，可同时运行：

| demo | tunnel | http/grpc | backend |
| --- | --- | --- | --- |
| quickstart | 17100 | 18100 | 18101 |
| auth | 17200 | 18200 | 18201 |
| grpc | 17300 | 19300 (grpc) | 19301 |
| tls | 17400 | 18400 | 18401 |

## 相关文档

- 配置与环境变量、CLI 用法见 [`../README.md`](../README.md)
- 安全模型与上线检查清单见 [`../SECURITY.md`](../SECURITY.md)
- P2P（ICE+QUIC）示例见 [`../../p2p/example`](../../p2p/example)
