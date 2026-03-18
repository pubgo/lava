# Lava 命令行文档（当前入口）

> 本文档以仓库根入口 `main.go` 为准。

## 命令总览

| 命令                    | 说明                       |
| ----------------------- | -------------------------- |
| `lava watch`            | 文件监听与自动执行命令     |
| `lava curl`             | Gateway 调试客户端         |
| `lava tunnel gateway`   | 启动 Tunnel Gateway        |
| `lava fileserver <dir>` | 本地目录静态文件服务       |
| `lava devproxy`         | 本地开发代理（DNS + HTTP） |

---

## 1. `lava watch`

文件变化监听器。读取配置文件（按优先级）：

1. `.lava/lava.yaml`
2. `.lava.yaml`
3. `lava.yaml`

若未找到配置，将使用内置默认 watcher。

### 核心字段

| 字段              | 说明                       |
| ----------------- | -------------------------- |
| `name`            | watcher 名称               |
| `directory`       | 监听目录                   |
| `patterns`        | 匹配模式（支持 `!` 排除）  |
| `commands`        | 文件变更后执行命令列表     |
| `ignore`          | 兼容字段，内部转为排除模式 |
| `ignore_patterns` | 兼容字段，内部转为排除模式 |
| `run_on_startup`  | 启动即执行一次             |
| `timeout`         | 单条命令超时（秒）         |

### 示例

```yaml
watch:
  watchers:
    - name: proto
      directory: .
      patterns:
        - "**/*.proto"
        - "!**/vendor"
      commands:
        - "protobuild gen"
      timeout: 30
```

---

## 2. `lava curl`

面向 Gateway 的轻量客户端，支持：

- operation 调用（如 `Service/Method`）
- 显式路径调用
- Header/Query/Path 参数注入
- `login` 子命令保存 token

### 常用参数

| 参数                     | 说明                      |
| ------------------------ | ------------------------- |
| `--addr`                 | 网关地址                  |
| `--prefix`               | 网关前缀（默认 `/api`）   |
| `--operation`            | 指定 operation            |
| `--path`                 | 显式路径                  |
| `-X, --method`           | 覆盖 HTTP 方法            |
| `-d, --data`             | 请求体字符串              |
| `--data-file`            | 从文件读取请求体          |
| `--stdin`                | 从标准输入读取请求体      |
| `-H, --header key=value` | 追加请求头                |
| `-Q, --query key=value`  | 追加 query                |
| `-P, --param key=value`  | 填充路径参数              |
| `--list`                 | 仅列出路由                |
| `--vars-name`            | debug vars 中网关信息名称 |
| `--timeout`              | 请求超时                  |
| `-k, --insecure`         | 跳过 TLS 校验             |
| `--raw`                  | 原样输出响应体            |

### 登录子命令

- `lava curl login -t <token>`
- `lava curl login --stdin`
- `lava curl login --env`（读取 `LAVA_TOKEN`）

Token 文件路径：`~/.lava/token`。

---

## 3. `lava tunnel gateway`

启动 Tunnel Gateway。

### 默认配置

| 配置项         | 默认值  | 对应环境变量         |
| -------------- | ------- | -------------------- |
| 监听地址       | `:7007` | `TUNNEL_LISTEN_ADDR` |
| HTTP 代理端口  | `8888`  | `TUNNEL_HTTP_PORT`   |
| gRPC 代理端口  | `9999`  | `TUNNEL_GRPC_PORT`   |
| Debug 代理端口 | `6066`  | `TUNNEL_DEBUG_PORT`  |
| 管理界面端口   | `:6067` | `TUNNEL_ADMIN_ADDR`  |

命令会同时启动：

- Gateway 服务
- Supervisor 管理与 debug
- 独立 debug 管理页面（默认 `:6067`）

---

## 4. `lava fileserver <dir>`

将指定目录作为静态文件服务对外暴露。

- 若未传 `<dir>`，默认使用当前工作目录。
- 端口使用运行时 HTTP 端口（通常为 `running.HttpPort`）。

常见用途：快速预览构建产物、临时共享静态目录。

---

## 5. `lava devproxy`

本地开发代理，提供：

- DNS 解析（`*.lava` -> `127.0.0.1`）
- HTTP 反向代理（按子域名匹配路由）

### 子命令

| 子命令      | 说明                  |
| ----------- | --------------------- |
| `start`     | 启动 devproxy         |
| `install`   | 安装系统集成（macOS） |
| `uninstall` | 卸载系统集成（macOS） |
| `routes`    | 输出当前路由          |

### 配置文件查找顺序

1. `.devproxy.json`
2. `.devproxy.yaml`
3. `.devproxy.yml`
4. `~/.devproxy.json`
5. `~/.devproxy.yaml`
6. `~/.devproxy.yml`

### 默认端口

- DNS：`5353`
- HTTP：`8080`

---

## 6. 说明：为何有些命令在代码里但跑不出来？

仓库中还存在一些命令包（如 `config/health/http/grpc/cron/version`），但它们并未注册到根入口 `main.go`。如果你从 `main.go` 构建的二进制执行，这些命令不会出现。

详见：`docs/modules/cmds.md`。
