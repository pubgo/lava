# 配置模型

Lava 服务示例（Tunnel Gateway、Scheduler 等）使用 [pubgo/funk](https://github.com/pubgo/funk) 的配置加载器，通过 YAML 分层合并与环境变量注入组装最终配置。

## 入口文件结构

典型入口配置（如 `internal/configs/tunnel.yaml`）：

```yaml
resources:
  - components

patch_resources:
  - .local.yaml

patch_envs:
  - envs
```

| 字段 | 含义 |
|------|------|
| `resources` | 要加载的资源目录或文件列表（相对入口 YAML 所在目录） |
| `patch_resources` | 在 resources 之后叠加的补丁文件（本地覆盖） |
| `patch_envs` | 环境变量定义目录（见下） |

## 合并顺序

以 `tunnel -c ./internal/configs/tunnel.yaml` 为例，加载顺序为：

1. **resources** — 合并 `internal/configs/components/` 下各组件 YAML（`tunnel.yaml`、`logger.yaml`、`http_server.yaml` 等），按文件名/声明顺序组合为一张配置图。
2. **patch_resources** — 合并 `internal/configs/.local.yaml`（若存在），用于开发者本机覆盖，通常不提交仓库。
3. **patch_envs** — 加载 `internal/configs/envs/` 目录中的环境定义，并将 `envs.yaml` 中的占位符替换为实际值；同时读取同目录或项目根下的 `.env`（若存在）。

后加载的层覆盖先加载的同名字段（深度合并由 funk `config` 包处理）。

## 目录布局

```
internal/configs/
├── tunnel.yaml              # 入口：声明 resources / patch
├── scheduler.yaml
├── components/              # 组件默认配置
│   ├── tunnel.yaml
│   ├── gateway_server.yaml
│   ├── logger.yaml
│   └── ...
├── envs/
│   ├── envs.yaml            # 环境键值（可被 .env 覆盖）
│   └── ...
└── .local.yaml              # 可选，本地补丁（gitignore）
```

## 在代码中指定配置路径

示例二进制在启动时设置入口配置：

```go
config.SetConfigPath("./internal/configs/tunnel.yaml")
```

`lavabuilder` 装配的服务通常通过 `-c` / `--config` 传入同一路径（见各 `cmds/*/cmd.go`）。

## 环境变量

- 组件配置中的 `${VAR}` 占位符由 `envs` 层解析。
- 运行时还可通过 `TUNNEL_*`、`HTTP_PORT` 等环境变量覆盖（见 `core/tunnel/config_env.go`、`core/running`）。
- Tunnel 生产环境应设置 `TUNNEL_AUTH_TOKEN`；`stage`/`prod` 下未设置且未声明 `TUNNEL_INSECURE=1` 时 Gateway 拒绝启动。

## 完整示例：Tunnel Gateway

```bash
# 1. 准备环境变量（可复制 envs/.env.example）
export TUNNEL_AUTH_TOKEN=dev-secret

# 2. 启动 Gateway
go run ./internal/examples/tunnel/main.go tunnel -c ./internal/configs/tunnel.yaml

# 3. 启动带 Agent 的 Scheduler
TUNNEL_GATEWAY_ADDR=localhost:7007 \
  go run ./internal/examples/scheduler/main.go cron -c ./internal/configs/scheduler.yaml
```

组件级字段说明见 `internal/configs/components/*.yaml` 内注释。

## 相关文档

- [quickstart.md](./quickstart.md) — 分步入门
- [development.md](./development.md) — Taskfile 与本地工作流
- [modules/core.md](./modules/core.md) — Supervisor 与生命周期
