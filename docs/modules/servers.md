# Servers 模块文档

`servers/*` 提供对外服务能力，当前主要包含 HTTP 与 gRPC 两类服务。

## 模块清单

| 模块            | 说明                                             | 关键入口                       |
| --------------- | ------------------------------------------------ | ------------------------------ |
| `servers/https` | Fiber HTTP 服务封装，默认接入 debug 与基础中间件 | `servers/https/server.go::New` |
| `servers/grpcs` | gRPC + Gateway 一体化服务，支持 gRPC/HTTP 双栈   | `servers/grpcs/server.go::New` |

## `servers/https` 要点

- 使用 Fiber 创建 HTTP 服务
- 默认中间件：serviceinfo / metric / accesslog / recovery
- 默认挂载：`/debug`
- 通过 `lava.HttpRouter` 进行路由装配

## `servers/grpcs` 要点

- 基于 `grpc.Server` 注册服务
- 同步注册到 `pkg/gateway.Mux`，支持 HTTP/JSON 调用
- 支持 `GrpcRouter` 与 `GrpcHttpRouter`
- 通过 `vars.Register` 暴露路由与服务信息

## 处理流程

```mermaid
flowchart LR
    Req1[HTTP Request] --> HTTPS[servers/https]
    Req2[gRPC Request] --> GRPCS[servers/grpcs]
    Req3[HTTP JSON] --> GW[pkg/gateway]
    GW --> GRPCS
    HTTPS --> Mid[Middlewares]
    GRPCS --> Mid
    Mid --> Biz[Business Handler]
```
