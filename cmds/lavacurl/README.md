# lavacurl

一个针对 Lava gateway 的轻量 HTTP 客户端，支持按 operation（gRPC 全方法名或自定义名称）或显式路径发起请求，并自动发现网关已注册的路由。

## 安装

```bash
go install ./cmds/lavacurl
```

## 快速上手

- 列出网关路由：
  ```bash
  lavacurl --list
  ```

- 按 operation 调用（自动匹配 method/path）：
  ```bash
  lavacurl Greeter/SayHello -d '{"name":"world"}'
  ```

- 按显式路径并指定方法：
  ```bash
  lavacurl --path /api/hello -X POST -d '{"name":"world"}'
  ```

- 携带 Header / Query / Path 参数：
  ```bash
  lavacurl Greeter/SayHello \
    -H "X-Req-Id=abc" \
    -Q verbose=true \
    -P id=42
  ```

## 常用参数

| 参数 | 说明 | 默认 |
| ---- | ---- | ---- |
| `--addr` | 网关地址 | `http://127.0.0.1:<HttpPort>` |
| `--prefix` | 网关前缀 | `/api` |
| `--operation` | operation 名（gRPC 全方法或自定义名） | - |
| `--path` | 显式路径（与 operation 二选一，也可用位置参数） | - |
| `-X, --method` | 覆盖 HTTP 方法 | GET/路由默认 |
| `-d, --data` | 请求体字符串 | - |
| `--data-file` | 从文件读取请求体 | - |
| `--stdin` | 从标准输入读取请求体 | - |
| `-H, --header key=value` | 追加 Header，可重复 | - |
| `-Q, --query key=value` | 追加 Query，可重复 | - |
| `-P, --param key=value` | 路径占位符填充，可重复 | - |
| `--timeout` | 请求超时 | `15s` |
| `-k, --insecure` | 跳过 TLS 校验 | `false` |
| `--raw` | 不做 JSON pretty-print | `false` |
| `--list` | 仅列出路由，不发请求 | `false` |
| `--vars-name` | gateway 路由信息的 expvar 名称 | `grpc-server-info` |

## 路由发现说明

`lavacurl` 默认向 `--addr` 的 `/debug/vars/api/list` 和 `/debug/vars/api/get/<vars-name>` 获取网关路由信息；若自定义了 expvar 名，可通过 `--vars-name` 指定。

## 返回输出

- 首行打印请求行，次行打印响应状态与耗时。
- 自动输出响应 Header。
- 若 `Content-Type` 含 `json` 且未指定 `--raw`，响应体将进行缩进格式化。

## 提示

- 位置参数可直接写 operation 或路径；`--path` / `--operation` 会覆盖位置参数。
- 路径参数形如 `/foo/{id}`，需通过 `-P id=123` 提供。
- gRPC-gateway 注册的路由也能被自动发现并调用。
