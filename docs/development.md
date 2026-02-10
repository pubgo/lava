# Lava 开发者指南

## 1. 开发环境搭建

### 1.1 系统要求

| 环境 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.25+ | 必须使用 Go 1.25 或更高版本 |
| Git | 2.0+ | 用于代码管理 |
| Task | 3.0+ | 用于运行构建和开发命令 |
| Protobuf | 3.0+ | 用于生成 gRPC 代码 |
| golangci-lint | 1.60+ | 用于代码 lint 检查 |
| goreleaser | 1.0+ | 用于发布版本（可选） |

### 1.2 安装开发依赖

#### 安装基础工具

```bash
# 安装 Go (macOS)
brew install go

# 安装 Task
brew install go-task/tap/go-task
eval "$(task --completion zsh)"

# 安装 Protobuf
brew install protobuf

# 安装 golangci-lint
brew install golangci-lint

# 安装 goreleaser (可选)
brew install goreleaser/tap/goreleaser
```

### 1.3 克隆代码库

```bash
# 克隆代码
git clone https://github.com/pubgo/lava.git
cd lava

# 安装依赖
go mod tidy

# 运行测试
task test

# 运行 lint 检查
task lint
```

## 2. 代码结构

### 2.1 目录结构

| 目录 | 说明 | 主要内容 |
|------|------|----------|
| `clients/` | 客户端库 | gRPC 和 HTTP 客户端 |
| `cmds/` | 命令行工具 | 各种命令行工具实现 |
| `core/` | 核心模块 | 服务管理、隧道、调度器等 |
| `docs/` | 文档 | 项目文档 |
| `internal/` | 内部实现 | 中间件、配置等 |
| `lava/` | 公共接口 | 核心接口定义 |
| `pkg/` | 公共包 | Gateway、工具函数等 |
| `proto/` | Protobuf 定义 | gRPC 服务定义 |
| `servers/` | 服务器实现 | HTTP 和 gRPC 服务器 |

### 2.2 核心模块说明

| 模块 | 位置 | 说明 |
|------|------|------|
| 服务管理 | `core/supervisor/` | 统一管理服务生命周期 |
| 隧道系统 | `core/tunnel/` | 服务代理和内网穿透 |
| 调度器 | `core/scheduler/` | 任务调度系统 |
| 日志系统 | `core/logging/` | 统一的日志接口 |
| 指标系统 | `core/metrics/` | 统一的指标接口 |
| 追踪系统 | `core/tracing/` | 链路追踪系统 |
| 编码系统 | `core/encoding/` | 统一的编码解码接口 |
| 服务发现 | `core/discovery/` | 服务发现系统 |

### 2.3 代码组织原则

1. **按功能模块组织**：相关功能放在同一目录下
2. **接口与实现分离**：核心接口放在 `lava/` 目录，实现放在相应模块中
3. **依赖注入**：使用 `dix` 依赖注入框架管理对象
4. **配置驱动**：通过配置文件控制服务行为
5. **中间件链**：使用中间件实现横切关注点

## 3. 开发流程

### 3.1 分支管理

| 分支 | 用途 | 说明 |
|------|------|------|
| `main` | 主分支 | 稳定版本，用于发布 |
| `develop` | 开发分支 | 集成新功能和修复 |
| `feature/*` | 特性分支 | 开发新功能 |
| `bugfix/*` | 修复分支 | 修复 bug |
| `release/*` | 发布分支 | 准备发布版本 |

### 3.2 开发步骤

1. **创建分支**：从 `develop` 分支创建新的特性分支
   ```bash
   git checkout develop
   git pull
   git checkout -b feature/my-feature
   ```

2. **实现功能**：按照编码规范实现新功能

3. **运行测试**：确保所有测试通过
   ```bash
   task test
   ```

4. **运行 lint**：确保代码符合 lint 规范
   ```bash
   task lint
   ```

5. **提交代码**：提交代码并推送到远程仓库
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/my-feature
   ```

6. **创建 PR**：在 GitHub 上创建 Pull Request

7. **代码审查**：等待代码审查和反馈

8. **合并代码**：代码审查通过后合并到 `develop` 分支

### 3.3 提交规范

使用统一的提交消息格式：

```bash
<type>(<scope>): <subject>

<body>

<footer>
```

**类型说明**：
- `feat`：新功能
- `fix`：bug 修复
- `docs`：文档更新
- `style`：代码风格调整
- `refactor`：代码重构
- `test`：测试相关
- `chore`：构建或依赖更新

**示例**：
```bash
feat(tunnel): add QUIC transport protocol

Add support for QUIC transport protocol in tunnel system

Closes #123
```

## 4. 编码规范

### 4.1 Go 编码规范

1. **遵循 Go 官方规范**：使用 `go fmt` 格式化代码
2. **命名规范**：
   - 包名：小写，使用单数形式
   - 函数名：驼峰命名，导出函数首字母大写
   - 变量名：驼峰命名，简短且有意义
   - 常量名：全大写，使用下划线分隔

3. **代码风格**：
   - 每行不超过 100 个字符
   - 使用 4 个空格缩进
   - 大括号放在行尾
   - 适当使用空行分隔代码块

4. **注释规范**：
   - 包级注释：每个包都要有包级注释
   - 函数注释：导出函数要有详细注释
   - 复杂代码：添加必要的注释说明

### 4.2 错误处理

1. **使用 pkg/errors**：使用 `github.com/pkg/errors` 包装错误
2. **错误信息**：错误信息要清晰、具体
3. **错误返回**：优先返回错误，而不是使用 panic
4. **错误检查**：所有错误都要检查和处理

**示例**：
```go
// 正确的错误处理
func GetUser(id string) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to query user")
    }
    return user, nil
}

// 错误的错误处理
func GetUser(id string) *User {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        panic(err) // 避免使用 panic
    }
    return user
}
```

### 4.3 测试规范

1. **测试覆盖率**：新代码要达到 80% 以上的测试覆盖率
2. **测试命名**：测试函数以 `Test` 开头
3. **测试文件**：测试文件以 `_test.go` 结尾
4. **测试隔离**：测试之间要相互隔离
5. **测试用例**：要覆盖正常、边界和错误场景

**示例**：
```go
func TestGetUser(t *testing.T) {
    // 正常场景
    user, err := GetUser("1")
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "1", user.ID)
    
    // 错误场景
    user, err = GetUser("")
    assert.Error(t, err)
    assert.Nil(t, user)
}
```

## 5. 测试指南

### 5.1 运行测试

```bash
# 运行所有测试
task test

# 运行特定包的测试
go test ./core/tunnel/...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 5.2 测试类型

| 测试类型 | 说明 | 示例 |
|----------|------|------|
| 单元测试 | 测试单个函数或方法 | `TestGetUser` |
| 集成测试 | 测试多个模块的交互 | `TestTunnelIntegration` |
| 端到端测试 | 测试完整的功能流程 | `TestFullWorkflow` |

### 5.3 测试工具

| 工具 | 用途 | 示例 |
|------|------|------|
| `testing` | Go 标准测试库 | `func TestXxx(t *testing.T)` |
| `testify` | 测试断言库 | `assert.Equal(t, expected, actual)` |
| `mockery` | 生成 mock 对象 | `mockery --name=Service` |
| `goconvey` | BDD 风格测试 | `So(user, ShouldNotBeNil)` |

### 5.4 测试最佳实践

1. **测试优先**：先写测试，再实现功能
2. **测试隔离**：使用 mock 或 stub 隔离外部依赖
3. **测试覆盖**：覆盖正常、边界和错误场景
4. **测试可读性**：测试代码要清晰、易读
5. **测试维护**：定期运行测试，确保测试通过

## 6. 插件开发

### 6.1 插件系统

Lava 采用插件化架构，支持通过插件扩展框架功能。插件可以是：

- **核心插件**：框架内置的插件
- **服务插件**：提供特定服务的插件
- **工具插件**：提供工具功能的插件
- **第三方插件**：由第三方开发的插件

### 6.2 开发插件

#### 6.2.1 中间件插件

**创建 HTTP/gRPC 中间件**：

```go
package mymiddleware

import (
    "context"
    "github.com/pubgo/lava/v2/lava"
)

// MyMiddleware 自定义中间件
type MyMiddleware struct{}

// New 创建中间件实例
func New() lava.Middleware {
    return &MyMiddleware{}
}

// Handle 处理请求
func (m *MyMiddleware) Handle(ctx context.Context, req interface{}) (interface{}, error) {
    // 前置处理
    // ...
    
    // 继续执行下一个中间件
    return req, nil
    
    // 后置处理
    // ...
}
```

**注册中间件**：

```go
// 在 main.go 中注册
httpService := https.New(https.Params{
    Handlers: []lava.HttpRouter{
        &UserRouter{},
    },
    Middlewares: []lava.Middleware{
        mymiddleware.New(),
    },
    // 其他参数...
})
```

#### 6.2.2 传输协议插件

**创建 Tunnel 传输协议**：

```go
package mytransport

import (
    "context"
    "github.com/pubgo/lava/v2/core/tunnel"
)

// MyTransport 自定义传输协议
type MyTransport struct{}

// init 自动注册传输协议
func init() {
    tunnel.RegisterTransport("mytransport", func(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
        return &MyTransport{}, nil
    })
}

// Name 传输协议名称
func (t *MyTransport) Name() string {
    return "mytransport"
}

// Dial 拨号连接
func (t *MyTransport) Dial(ctx context.Context, addr string) (tunnel.Session, error) {
    // 实现拨号逻辑
    // ...
    return &MySession{}, nil
}

// Listen 监听连接
func (t *MyTransport) Listen(ctx context.Context, addr string) (tunnel.Listener, error) {
    // 实现监听逻辑
    // ...
    return &MyListener{}, nil
}
```

**使用自定义传输协议**：

```go
// 在配置中使用
cfg := &tunnelagent.Config{
    GatewayAddr: "localhost:7000",
    Transport:   "mytransport", // 使用自定义传输协议
    // 其他配置...
}
```

#### 6.2.3 服务发现插件

**创建服务发现实现**：

```go
package mydiscovery

import (
    "context"
    "github.com/pubgo/lava/v2/core/discovery"
)

// MyDiscovery 自定义服务发现
type MyDiscovery struct{}

// New 创建服务发现实例
func New(cfg *discovery.Config) discovery.Discovery {
    return &MyDiscovery{}
}

// Register 注册服务
func (d *MyDiscovery) Register(ctx context.Context, service *discovery.ServiceInfo) error {
    // 实现注册逻辑
    // ...
    return nil
}

// Deregister 注销服务
func (d *MyDiscovery) Deregister(ctx context.Context, service *discovery.ServiceInfo) error {
    // 实现注销逻辑
    // ...
    return nil
}

// Discover 发现服务
func (d *MyDiscovery) Discover(ctx context.Context, serviceName string) ([]*discovery.ServiceInstance, error) {
    // 实现发现逻辑
    // ...
    return []*discovery.ServiceInstance{}, nil
}

// Watch 监控服务变化
func (d *MyDiscovery) Watch(ctx context.Context, serviceName string, callback func([]*discovery.ServiceInstance)) error {
    // 实现监控逻辑
    // ...
    return nil
}
```

### 6.3 插件注册

#### 6.3.1 自动注册

使用 `init()` 函数自动注册插件：

```go
// 在插件包中
func init() {
    // 注册中间件
    lava.RegisterMiddleware("mymiddleware", mymiddleware.New)
    
    // 注册传输协议
    tunnel.RegisterTransport("mytransport", func(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
        return &MyTransport{}, nil
    })
    
    // 注册服务发现
    discovery.RegisterDiscovery("mydiscovery", mydiscovery.New)
}
```

#### 6.3.2 手动注册

在应用启动时手动注册插件：

```go
func main() {
    // 手动注册中间件
    lava.RegisterMiddleware("mymiddleware", mymiddleware.New)
    
    // 创建依赖注入容器
    di := lavabuilder.New()
    
    // 运行应用
    lavabuilder.Run(di)
}
```

## 7. 命令行工具开发

### 7.1 创建命令

**创建新命令**：

```go
package mycmd

import (
    "github.com/pubgo/lava/v2/pkg/cliutil"
    "github.com/redantlabs/redant"
)

// New 创建命令实例
func New() *redant.Command {
    cmd := cliutil.NewCommand(
        "mycmd",
        "My custom command",
        func(cmd *redant.Command) error {
            // 命令实现
            name := cmd.Flag("name").String()
            cmd.Println("Hello, ", name)
            return nil
        },
    )
    
    // 添加命令行参数
    cmd.Flag("name", "Your name").Default("World").String()
    
    return cmd
}
```

**注册命令**：

```go
// 在 lavabuilder/builder.go 中注册
func Run(di *dix.Dix) {
    // ...
    
    // 注册命令
    dix.Provide(di, mycmd.New)
    
    // ...
}
```

### 7.2 命令结构

**命令结构**：

```go
// 命令结构
cmd := &redant.Command{
    Use:   "mycmd",           // 命令名称
    Short: "Short description", // 简短描述
    Long:  "Long description",  // 详细描述
    Run: func(cmd *redant.Command) error {
        // 命令实现
        return nil
    },
    Flags: []redant.Flag{
        // 命令行参数
    },
    Children: []*redant.Command{
        // 子命令
    },
}
```

**示例**：
```go
// 创建带有子命令的命令
parentCmd := cliutil.NewCommand(
    "parent",
    "Parent command",
    func(cmd *redant.Command) error {
        cmd.Println("Parent command")
        return nil
    },
)

childCmd := cliutil.NewCommand(
    "child",
    "Child command",
    func(cmd *redant.Command) error {
        cmd.Println("Child command")
        return nil
    },
)

parentCmd.AddCommand(childCmd)
```

## 8. 扩展核心功能

### 8.1 扩展服务类型

**创建自定义服务**：

```go
package myservice

import (
    "context"
    "github.com/pubgo/lava/v2/core/supervisor"
)

// MyService 自定义服务
type MyService struct{}

// String 服务名称
func (s *MyService) String() string {
    return "my-service"
}

// Serve 服务启动逻辑
func (s *MyService) Serve(ctx context.Context) error {
    // 服务实现
    // ...
    
    <-ctx.Done()
    return nil
}

// New 创建服务实例
func New() supervisor.Service {
    return supervisor.NewService("my-service", func(ctx context.Context) error {
        service := &MyService{}
        return service.Serve(ctx)
    })
}
```

**注册服务**：

```go
func main() {
    di := lavabuilder.New()
    
    // 注册自定义服务
    di.Provide(func() supervisor.Service {
        return myservice.New()
    })
    
    lavabuilder.Run(di)
}
```

### 8.2 扩展配置系统

**创建自定义配置源**：

```go
package myconfig

import (
    "github.com/pubgo/lava/v2/core/config"
)

// MyConfigSource 自定义配置源
type MyConfigSource struct{}

// New 创建配置源实例
func New() config.ConfigSource {
    return &MyConfigSource{}
}

// Load 加载配置
func (s *MyConfigSource) Load() (map[string]interface{}, error) {
    // 从自定义源加载配置
    // ...
    return map[string]interface{}{},
        nil
}

// Watch 监控配置变化
func (s *MyConfigSource) Watch(callback func(map[string]interface{})) error {
    // 监控配置变化
    // ...
    return nil
}
```

**注册配置源**：

```go
func init() {
    config.RegisterConfigSource("myconfig", myconfig.New)
}
```

### 8.3 扩展日志系统

**创建自定义日志输出**：

```go
package mylogger

import (
    "github.com/pubgo/lava/v2/core/logging"
)

// MyLogger 自定义日志输出
type MyLogger struct{}

// New 创建日志实例
func New(cfg *logging.Config) logging.Logger {
    return &MyLogger{}
}

// Info 信息日志
func (l *MyLogger) Info() logging.Logger {
    // 实现信息日志
    return l
}

// Debug 调试日志
func (l *MyLogger) Debug() logging.Logger {
    // 实现调试日志
    return l
}

// Error 错误日志
func (l *MyLogger) Error() logging.Logger {
    // 实现错误日志
    return l
}

// Msg 输出日志消息
func (l *MyLogger) Msg(msg string) {
    // 输出日志消息
    println(msg)
}

// Str 添加字符串字段
func (l *MyLogger) Str(key string, value string) logging.Logger {
    // 添加字符串字段
    return l
}
```

**注册日志输出**：

```go
func init() {
    logging.RegisterLogger("mylogger", mylogger.New)
}
```

## 9. 贡献代码

### 9.1 贡献流程

1. **Fork 仓库**：在 GitHub 上 fork 代码库
2. **创建分支**：从 `develop` 分支创建特性分支
3. **实现功能**：按照编码规范实现功能
4. **运行测试**：确保所有测试通过
5. **运行 lint**：确保代码符合 lint 规范
6. **提交代码**：使用规范的提交消息
7. **创建 PR**：创建 Pull Request 到 `develop` 分支
8. **代码审查**：等待代码审查和反馈
9. **合并代码**：代码审查通过后合并到 `develop` 分支

### 9.2 贡献指南

1. **Issue 管理**：
   - 在实现功能前先创建 Issue
   - 参考相关 Issue 在提交消息中
   - 关闭相关 Issue 在合并后

2. **代码质量**：
   - 确保代码测试覆盖率达到 80% 以上
   - 确保代码通过 lint 检查
   - 确保代码符合 Go 编码规范

3. **文档更新**：
   - 为新功能更新文档
   - 为 API 变更更新文档
   - 确保文档与代码同步

4. **版本兼容性**：
   - 保持向后兼容性
   - 对于破坏性变更，在文档中说明
   - 提供迁移指南

### 9.3 代码审查

**代码审查要点**：

1. **功能正确性**：代码是否实现了预期功能
2. **代码质量**：代码是否符合编码规范
3. **测试覆盖**：是否有足够的测试覆盖
4. **性能考虑**：代码是否有性能问题
5. **安全考虑**：代码是否有安全问题
6. **文档完整性**：是否有必要的文档

**代码审查流程**：
1. **自我审查**：提交前先自我审查代码
2. **同伴审查**：邀请团队成员审查代码
3. **修复问题**：根据审查意见修复问题
4. **重新提交**：修复后重新提交代码
5. **最终审查**：确认所有问题都已解决

## 10. 发布流程

### 10.1 版本管理

使用语义化版本管理：

```
MAJOR.MINOR.PATCH
```

- **MAJOR**：破坏性变更
- **MINOR**：新功能
- **PATCH**：bug 修复

### 10.2 发布准备

**发布前检查**：

1. **运行测试**：确保所有测试通过
2. **运行 lint**：确保代码符合 lint 规范
3. **更新版本**：更新 VERSION 文件
4. **更新文档**：更新发布说明和文档
5. **检查依赖**：确保依赖版本正确

**更新版本**：

```bash
# 更新版本号
echo "1.2.3" > .version/VERSION

# 提交版本更新
git add .version/VERSION
git commit -m "chore: bump version to 1.2.3"
```

### 10.3 构建发布

**使用 goreleaser**：

```bash
# 配置 .goreleaser.yaml
# ...

# 构建发布
goreleaser release --snapshot --rm-dist

# 正式发布
goreleaser release --rm-dist
```

**手动构建**：

```bash
# 构建不同平台的二进制文件
go build -o ./bin/lava-darwin-amd64 ./main.go
go build -o ./bin/lava-linux-amd64 ./main.go
go build -o ./bin/lava-windows-amd64.exe ./main.go

# 打包发布文件
zip -r lava-1.2.3-darwin-amd64.zip ./bin/lava-darwin-amd64
zip -r lava-1.2.3-linux-amd64.zip ./bin/lava-linux-amd64
zip -r lava-1.2.3-windows-amd64.zip ./bin/lava-windows-amd64.exe
```

### 10.4 发布说明

**创建发布说明**：

```markdown
# Release v1.2.3

## Features

- Add QUIC transport protocol support
- Add new command line tool `lavacurl`
- Improve error handling in tunnel system

## Fixes

- Fix memory leak in gRPC server
- Fix race condition in supervisor
- Fix configuration loading issue

## Breaking Changes

- None

## Documentation

- Update architecture documentation
- Add quickstart guide
- Update API reference
```

## 11. 开发工具

### 11.1 常用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `task test` | 运行测试 | `task test` |
| `task lint` | 运行 lint 检查 | `task lint` |
| `task build` | 构建项目 | `task build` |
| `task clean` | 清理构建产物 | `task clean` |
| `task fmt` | 格式化代码 | `task fmt` |
| `task proto` | 生成 Protobuf 代码 | `task proto` |

### 11.2 开发脚本

**Task 配置**：

```yaml
# taskfile.yml
version: '3'

vars:
  GoVersion: "1.25"
  ProjectName: "lava"

 tasks:
  default:
    cmds:
      - task: test

  test:
    desc: "Run tests"
    cmds:
      - go test ./... -timeout 30s

  lint:
    desc: "Run lint checks"
    cmds:
      - golangci-lint run

  build:
    desc: "Build project"
    cmds:
      - go build -o ./bin/{{.ProjectName}} ./main.go

  clean:
    desc: "Clean build artifacts"
    cmds:
      - rm -rf ./bin
      - rm -rf ./dist

  fmt:
    desc: "Format code"
    cmds:
      - go fmt ./...

  proto:
    desc: "Generate protobuf code"
    cmds:
      - protoc --go_out=./proto --go-grpc_out=./proto ./proto/*.proto
```

### 11.3 IDE 配置

**VS Code 配置**：

```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v"],
  "go.buildTags": [],
  "go.useLanguageServer": true,
  "[go]": {
    "editor.tabSize": 4,
    "editor.insertSpaces": true,
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

**GoLand 配置**：

- **Go 版本**：设置为 1.25+  
- **格式化工具**：使用 `gofmt`
- **Lint 工具**：使用 `golangci-lint`
- **测试配置**：启用详细输出
- **代码风格**：使用 4 空格缩进

## 12. 常见问题

### 12.1 开发问题

**问题**：依赖解析失败

**解决方案**：
```bash
# 清理依赖缓存
go clean -modcache

# 重新生成 go.mod
go mod tidy
```

**问题**：测试失败

**解决方案**：
```bash
# 运行单个测试
go test -v ./core/tunnel -run TestTunnel

# 查看测试覆盖率
go test -coverprofile=coverage.out ./core/tunnel
```

**问题**：lint 检查失败

**解决方案**：
```bash
# 运行单个文件的 lint 检查
golangci-lint run ./core/tunnel/...

# 自动修复 lint 问题
golangci-lint run --fix
```

### 12.2 构建问题

**问题**：构建失败

**解决方案**：
```bash
# 检查 Go 版本
go version

# 检查依赖
go mod tidy

# 清理并重新构建
task clean
task build
```

**问题**：交叉编译失败

**解决方案**：
```bash
# 设置交叉编译环境变量
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# 构建
go build -o ./bin/lava-linux-amd64 ./main.go
```

### 12.3 发布问题

**问题**：goreleaser 失败

**解决方案**：
```bash
# 检查 goreleaser 配置
cat .goreleaser.yaml

# 运行快照构建
goreleaser release --snapshot --rm-dist

# 检查 Git 标签
git tag
```

**问题**：版本号冲突

**解决方案**：
```bash
# 检查当前版本
echo $(cat .version/VERSION)

# 检查 Git 标签
git tag -l

# 创建新标签
git tag v1.2.3
git push origin v1.2.3
```

## 13. 总结

Lava 框架提供了一套完整的微服务开发解决方案，通过统一的抽象和丰富的功能，大大简化了微服务的开发和运维。作为开发者，你可以：

1. **贡献代码**：为 Lava 框架贡献新功能和 bug 修复
2. **开发插件**：开发自定义中间件、传输协议等插件
3. **扩展功能**：扩展核心功能，如服务类型、配置系统等
4. **构建应用**：使用 Lava 框架构建自己的微服务应用

通过遵循本开发者指南，你可以：
- 快速搭建开发环境
- 了解代码结构和开发流程
- 遵循编码规范和最佳实践
- 开发高质量的插件和扩展
- 贡献代码并参与社区建设

Lava 框架欢迎所有开发者的贡献和反馈，让我们一起构建更好的微服务开发体验！