# Lava 接口层模块（`lava/*`）

`lava/*` 定义跨模块共享的核心接口，是框架的抽象边界。

## 文件与职责

| 文件                 | 职责                                       |
| -------------------- | ------------------------------------------ |
| `lava/middleware.go` | 中间件与链式组合抽象                       |
| `lava/request.go`    | 请求抽象接口                               |
| `lava/response.go`   | 响应抽象接口                               |
| `lava/router.go`     | HTTP/gRPC Router 接口定义                  |
| `lava/server.go`     | 通用 server/closer/listener/validator 接口 |

## 关键价值

1. 降低模块耦合：`servers/*`、`clients/*` 通过接口协作。
2. 保证扩展一致性：新增组件时优先对齐这层接口。
3. 让中间件在不同协议场景下复用。

## 推荐实践

- 新增跨模块抽象时先评估是否应放入 `lava/*`。
- 避免在接口层引入业务依赖。
- 文档示例优先引用真实签名，防止“接口漂移”。
