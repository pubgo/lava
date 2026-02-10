# lava 命令行工具

`lava` 命令是 Lava 框架的统一命令行入口，提供了多种子命令来帮助开发者更高效地使用和开发基于 Lava 框架的应用。

## 功能特性

- **统一入口**：所有命令通过 `lava` 命令作为统一入口，使用更加方便
- **文件监控**：通过 `watch` 子命令监控文件变更并自动执行构建命令
- **gRPC 客户端**：通过 `curl` 子命令向 gRPC 服务发送 HTTP 请求
- **自动构建**：当文件变更时自动执行相应的构建命令
- **路由发现**：`curl` 命令支持自动发现网关已注册的路由

## 安装

### 从源码构建

```bash
# 克隆代码库
git clone https://github.com/pubgo/lava.git
cd lava

# 构建二进制文件
go build -o lava .

# 将二进制文件添加到 PATH（可选）
cp lava /usr/local/bin/
```

## 基本使用

### 查看帮助信息

```bash
lava
```

这将显示 `lava` 命令的基本信息和可用的子命令。

## 子命令

### 1. watch 命令

**功能**：监控文件变更并自动执行相应的构建命令。支持配置多个 watcher，每个 watcher 可以监控不同的目录和执行不同的命令。

**用法**：
```bash
lava watch
```

**配置文件**：
`watch` 命令通过配置文件来定义多个 watcher，配置文件支持以下位置（按优先级从高到低）：
1. `.lava/lava.yaml`
2. `.lava.yaml`
3. `lava.yaml`

**配置格式**：
```yaml
watch:
  watchers:
    # watcher 1：监控 proto 文件
    - name: "proto"
      directory: "."
      patterns:
        - "**/*.proto"
        - "!**/vendor"
        - "!**/.git"
        - "!*.tmp"
        - "!*~"
      commands:
        - "protobuild gen"
      run_on_startup: false
      timeout: 30

    # watcher 2：监控 go 文件
    - name: "go"
      directory: "."
      patterns:
        - "**/*.go"
        - "!**/dist"
        - "!**/build"
        - "!**/vendor"
        - "!**/node_modules"
        - "!**/.git"
        - "!*.tmp"
        - "!*~"
        - "!.DS_Store"
      commands:
        - "go build ./..."
      run_on_startup: false
      timeout: 30
```

**配置参数说明**：

| 参数 | 说明 | 必填 | 默认值 |
|------|------|------|--------|
| `name` | watcher 名称 | 是 | - |
| `directory` | 要监控的目录 | 是 | - |
| `patterns` | 文件匹配模式列表 (包含：`**/*.go`，排除：`!**/vendor`) | 否 | `["*"]` |
| `commands` | 文件变更后执行的命令列表 | 是 | - |
| `run_on_startup` | 是否在启动时执行一次命令 | 否 | `false` |
| `timeout` | 命令执行的超时时间（秒） | 否 | `30` |

**示例**：
```bash
# 使用默认配置运行（监控当前目录下的 .proto 和 .go 文件）
lava watch

# 使用自定义配置文件运行
# 创建 .lava.yaml 文件并配置多个 watcher
lava watch
```

**工作原理**：
1. 加载配置文件，支持配置多个 watcher
2. 每个 watcher 在独立的 goroutine 中运行，可以同时监控不同的目录
3. 当文件发生变更时，根据文件类型执行相应的构建命令
4. 自动忽略 `.git`、`node_modules`、`vendor` 等不需要监控的目录
5. 当创建新目录时，自动将其添加到监控列表中
6. 支持命令执行超时，超时后会自动终止命令进程

### 2. curl 命令

**功能**：一个针对 Lava gateway 的轻量 HTTP 客户端，支持按 operation（gRPC 全方法名或自定义名称）或显式路径发起请求，并自动发现网关已注册的路由。

**用法**：
```bash
lava curl [options] [operation/path]
```

**常用参数**：

| 参数 | 说明 | 默认值 |
|------|------|--------|
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

**示例**：

#### 列出网关路由

```bash
lava curl --list
```

#### 登录保存 Token

Token 将保存在 `~/.lava/curl/token`（`0600` 权限），后续请求若未显式设置 `Authorization`，会自动注入 `Bearer <token>`。

```bash
# 直接提供 token
lava curl login -t "YOUR_TOKEN"

# 从 stdin 读取 token
echo -n "YOUR_TOKEN" | lava curl login --stdin

# 从环境变量读取 token
curl_TOKEN="YOUR_TOKEN" lava curl login --env
```

#### 按 operation 调用

自动匹配 method/path：

```bash
lava curl Greeter/SayHello -d '{"name":"world"}'
```

#### 按显式路径并指定方法

```bash
lava curl --path /api/hello -X POST -d '{"name":"world"}'
```

#### 携带 Header / Query / Path 参数

```bash
lava curl Greeter/SayHello \
  -H "X-Req-Id=abc" \
  -Q verbose=true \
  -P id=42
```

#### 从文件读取请求体

```bash
lava curl Greeter/SayHello --data-file ./request.json
```

#### 从标准输入读取请求体

```bash
echo '{"name":"world"}' | lava curl Greeter/SayHello --stdin
```

**路由发现说明**：

`curl` 默认向 `--addr` 的 `/debug/vars/api/list` 和 `/debug/vars/api/get/<vars-name>` 获取网关路由信息；若自定义了 expvar 名，可通过 `--vars-name` 指定。

**返回输出**：

- 首行打印请求行，次行打印响应状态与耗时
- 自动输出响应 Header
- 若 `Content-Type` 含 `json` 且未指定 `--raw`，响应体将进行缩进格式化

## 配置

### 配置文件位置

`lava` 命令支持配置文件，配置文件可以放在以下位置（按优先级从高到低）：

1. `.lava/lava.yaml`：优先级最高
2. `.lava.yaml`：优先级次之
3. `lava.yaml`：优先级最低

当多个配置文件存在时，优先级高的配置文件会覆盖优先级低的配置文件。

### 配置文件格式

配置文件使用 YAML 格式，包含 `watch` 和 `curl` 两个主要配置部分。

**示例配置文件**：

```yaml
# watch 命令配置
watch:
  # watcher 列表，可以配置多个 watcher
  watchers:
    # watcher 1：监控 proto 文件
    - name: "proto"
      directory: "."
      patterns:
        - "**/*.proto"
        - "!**/vendor"
        - "!**/.git"
        - "!*.tmp"
        - "!*~"
      commands:
        - "protobuild gen"
      run_on_startup: false
      timeout: 30

    # watcher 2：监控 go 文件
    - name: "go"
      directory: "."
      patterns:
        - "**/*.go"
        - "!**/dist"
        - "!**/build"
        - "!**/vendor"
        - "!**/node_modules"
        - "!**/.git"
        - "!*.tmp"
        - "!*~"
        - "!.DS_Store"
      commands:
        - "go build ./..."
      run_on_startup: false
      timeout: 30

# curl 命令配置
curl:
  # 网关地址
  addr: "http://127.0.0.1:8080"
  
  # 网关前缀
  prefix: "/api"
  
  # 请求超时时间
  timeout: "15s"
  
  # 是否跳过 TLS 校验
  insecure: false
  
  # 是否对 JSON 响应进行格式化
  pretty: true
  
  # gateway 路由信息的 expvar 名称
  vars_name: "grpc-server-info"
  
  # 默认的请求头
  headers:
    - "Content-Type: application/json"
  
  # 默认的查询参数
  queries: {}
```

### 环境变量

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `LAVA_TOKEN` | `lava curl login --env` 时使用的 token | - |

### 配置文件与命令行参数的优先级

配置文件中的设置可以被命令行参数覆盖，优先级如下：

1. 命令行参数（优先级最高）
2. 配置文件
3. 默认值（优先级最低）

**示例**：

```bash
# 配置文件中设置了 addr 为 http://127.0.0.1:8080
# 但命令行参数会覆盖配置文件的设置
lava curl --addr http://localhost:9090 Greeter/SayHello
```

### 配置文件示例

项目根目录中提供了一个示例配置文件 `.lava.yaml.example`，可以参考该文件创建自己的配置文件：

```bash
# 复制示例配置文件
cp .lava.yaml.example .lava.yaml

# 根据需要修改配置文件
vim .lava.yaml
```

## 常见问题

### 1. watch 命令不工作

**可能原因**：
- 文件系统不支持 inotify（如某些网络文件系统）
- 监控的文件数量超过了系统限制

**解决方案**：
- 尝试监控更小的目录范围
- 检查系统 inotify 限制并调整：
  ```bash
  # 查看当前限制
  cat /proc/sys/fs/inotify/max_user_watches
  
  # 临时调整限制
  sudo sysctl fs.inotify.max_user_watches=524288
  
  # 永久调整限制
  echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
  sudo sysctl -p
  ```

### 2. curl 无法发现路由

**可能原因**：
- 网关地址不正确
- 网关未启用调试接口
- 路由信息的 expvar 名称不正确

**解决方案**：
- 确认网关地址正确：`lava curl --addr http://localhost:8080 --list`
- 确认网关已启用调试接口
- 尝试指定正确的 expvar 名称：`lava curl --vars-name custom-vars-name --list`

### 3. 构建命令执行失败

**可能原因**：
- 构建环境配置不正确
- 代码存在错误
- 依赖缺失

**解决方案**：
- 手动执行构建命令查看详细错误信息
- 检查代码是否存在语法错误或逻辑错误
- 运行 `go mod tidy` 确保依赖正确

## 最佳实践

### 1. 开发时使用 watch 命令

在开发过程中，使用 `watch` 命令监控文件变更并自动执行构建命令，可以大大提高开发效率：

```bash
# 在一个终端中运行 watch 命令
lava watch

# 在另一个终端中进行开发
# 当文件变更时，watch 命令会自动执行构建命令
```

### 2. 使用 curl 测试 gRPC 服务

在开发和测试 gRPC 服务时，使用 `curl` 命令可以更方便地向服务发送请求：

```bash
# 列出所有可用的路由
lava curl --list

# 测试特定的接口
lava curl Greeter/SayHello -d '{"name":"test"}'
```

### 3. 结合 CI/CD 使用

在 CI/CD 流程中，可以使用 `lava` 命令来执行构建和测试：

```bash
# 构建项目
lava watch --once # 只执行一次构建命令

# 测试服务
lava curl Health/Check
```

## 示例

### 示例 1：监控 proto 文件变更

```bash
# 监控 proto 目录下的文件变更
lava watch ./proto

# 当 proto 文件变更时，会自动执行 protobuild gen 命令
```

### 示例 2：测试 gRPC 服务

```bash
# 启动 gRPC 服务（假设服务运行在 localhost:8080）
# ...

# 列出所有可用的路由
lava curl --addr http://localhost:8080 --list

# 测试 SayHello 接口
lava curl --addr http://localhost:8080 Greeter/SayHello -d '{"name":"world"}'
```

## 总结

`lava` 命令是 Lava 框架的强大命令行工具，通过提供统一的命令入口和多种实用的子命令，大大简化了开发者的工作流程。无论是监控文件变更、自动构建，还是测试 gRPC 服务，`lava` 命令都能提供便捷的解决方案。

通过合理使用 `lava` 命令，开发者可以：
- 提高开发效率，减少手动执行构建命令的次数
- 更方便地测试和调试 gRPC 服务
- 简化 CI/CD 流程中的构建和测试步骤

`lava` 命令是 Lava 框架生态系统中的重要组成部分，为开发者提供了一站式的开发工具解决方案。