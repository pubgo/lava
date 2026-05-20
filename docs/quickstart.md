# Lava 快速开始

本文档面向“先跑起来，再看源码”。

## 1. 环境准备

- Go `1.25+`
- Task

macOS 示例：

```bash
brew install go
brew install go-task/tap/go-task
```

## 2. 获取并初始化

```bash
git clone https://github.com/pubgo/lava.git
cd lava
go mod tidy
```

## 3. 基础质量检查

```bash
task test
task lint
```

## 4. 构建 CLI

```bash
go build -o lava .
./lava
```

当前根入口可用命令：

- `watch`
- `curl`
- `tunnel gateway`
- `fileserver <dir>`
- `devproxy`

## 5. 常用体验路径

### 5.1 本地文件服务

```bash
./lava fileserver .
```

然后访问输出日志中的本地地址。

### 5.2 文件变更监听

```bash
./lava watch
```

默认会尝试读取以下配置文件：

1. `.lava/lava.yaml`
2. `.lava.yaml`
3. `lava.yaml`

找不到时使用内置默认 watcher。

### 5.3 Tunnel Gateway

```bash
./lava tunnel gateway
```

可通过环境变量覆盖端口和地址：

- `TUNNEL_LISTEN_ADDR`
- `TUNNEL_HTTP_PORT`
- `TUNNEL_GRPC_PORT`
- `TUNNEL_DEBUG_PORT`
- `TUNNEL_ADMIN_ADDR`

### 5.4 devproxy

```bash
./lava devproxy start
```

macOS 可选安装系统集成：

```bash
sudo ./lava devproxy install
```

## 6. Proto 工作流

```bash
task proto:fmt
task proto:lint
task proto:gen
```

相关配置：`protobuf.yaml`。

## 7. 下一步阅读

- 架构：`docs/architecture-v2.md`
- 设计：`docs/design-v2.md`
- 模块：`docs/modules/README.md`
- 命令：`docs/lava-command.md`
