# Lava 文档总览

这套文档以当前仓库实现为准，重点覆盖：

- 架构与运行流程
- 设计原则与核心抽象
- 各模块职责与入口
- CLI 使用方式

## 推荐阅读顺序

1. `quickstart.md`：分步入门（推荐新同学从这里开始）
2. `development.md`：Taskfile、测试与 FAQ
3. `configuration.md`：resources / patch / envs 配置模型
4. `architecture-v2.md`：全局架构认知
5. `design-v2.md`：抽象与设计取舍
6. `modules/README.md`：按目录定位模块
7. `lava-command.md`：查命令参数与示例
8. `zrpc.md`：NATS + protobuf unary RPC

## 文档地图

- 入门：`quickstart.md`、`development.md`、`configuration.md`
- 架构：`architecture-v2.md`（分层、四条数据通路、gatewayserver / zrpcbridge / 部署）
- 设计：`design-v2.md`
- P2P：`design-p2p.md`
- zrpc：`zrpc.md`
- 命令：`lava-command.md`
- Supervisor：`supervisor.md`
- Copilot：`copilot-skills.md`
- Gateway 部署（TLS / HTTP/3）：`../pkg/gateway/docs/deploy.md`
- Traefik 示例：`../deploy/traefik/README.md`
- 模块：`modules/README.md`
  - `modules/core.md`
  - `modules/servers.md`
  - `modules/clients.md`
  - `modules/pkg.md`
  - `modules/cmds.md`
  - `modules/lava.md`
  - `modules/internal.md`

## 约定说明

- 文档中的路径均为仓库相对路径。
- 命令示例默认在仓库根目录执行。
- 若文档与代码不一致，以代码为准，欢迎提交 PR 修正文档。
