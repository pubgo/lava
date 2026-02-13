# Supervisor 服务管理模块

## 概述
`core/supervisor` 提供统一的服务生命周期管理能力：注册服务、启动/停止/重启、自动重启策略、状态与指标查询，并配套调试 UI/API。

该模块适合在微服务进程内统一托管 HTTP/gRPC/后台任务等服务，实现可靠的启动顺序、失败恢复和运行可观测性。

## 核心概念

- **Manager**：服务管理器，维护服务集合与生命周期控制。
- **Service**：被管理的服务，需实现 `Serve(ctx)` 与 `Metric()` 等接口。
- **ServiceConfig**：服务级别配置（重启策略、退避、窗口期等）。
- **Option**：函数式配置选项，便于按需覆盖默认配置。
- **ServiceInfo**：服务运行态信息（重启次数、失败状态、窗口期统计等）。

## 快速开始

### 1) 定义服务

```go
import (
    "context"
    "github.com/pubgo/lava/v2/core/supervisor"
)

func mainTask(ctx context.Context) error {
    // 你的业务逻辑
    <-ctx.Done()
    return nil
}

srv := supervisor.NewService("main-task", mainTask)
```

或实现 `Service` 接口：

```go
type MyService struct{}

func (s *MyService) Name() string { return "my-service" }
func (s *MyService) Error() error { return nil }
func (s *MyService) String() string { return s.Name() }
func (s *MyService) Metric() *supervisor.Metric { return &supervisor.Metric{Name: s.Name()} }
func (s *MyService) Serve(ctx context.Context) error {
    <-ctx.Done()
    return nil
}
```

### 2) 创建 Manager 并注册服务

```go
import (
    "context"
    "github.com/pubgo/lava/v2/core/lifecycle"
    "github.com/pubgo/lava/v2/core/supervisor"
)

lc := lifecycle.New()
manager := supervisor.NewManager("app", lc)

// 默认自动启动
_ = manager.Add(srv)

// 不自动启动，按需手动启动
_ = manager.Add(srv, supervisor.WithAutoStart(false))
```

### 3) 运行 Manager

```go
ctx, cancel := context.WithCancel(context.Background())

defer cancel()
_ = manager.Run(ctx) // 启动并阻塞，直到 ctx 被取消
```

## 主要 API

### Manager

```go
func NewManager(name string, lc lifecycle.Getter) *Manager
func Default(lc lifecycle.Getter) *Manager

func (m *Manager) Add(srv Service, opts ...Option) error
func (m *Manager) AddWithConfig(srv Service, config ServiceConfig) error
func (m *Manager) Delete(name string) error
func (m *Manager) RemoveServices() error

func (m *Manager) StartService(name string) error
func (m *Manager) StopService(name string) error
func (m *Manager) RestartService(name string) error
func (m *Manager) RestartServices() error
func (m *Manager) ResetService(name string) error

func (m *Manager) GetServicesInfo() []*ServiceInfo
func (m *Manager) GetServiceInfo(name string) (*ServiceInfo, error)
func (m *Manager) Services() []Service

func (m *Manager) Run(ctx context.Context) error
func (m *Manager) Serve(ctx context.Context) error
func (m *Manager) ServeBackground(ctx context.Context) <-chan error
```

### ServiceConfig 与 Option

`ServiceConfig` 控制服务重启策略与退避行为。`Add` 会使用默认配置并应用 `Option`。

默认值：

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| `AutoStart` | `true` | 是否自动启动 |
| `RestartPolicy` | `RestartAlways` | 重启策略 |
| `MaxRestarts` | `0` | 总重启次数上限，0 表示无限制 |
| `RestartDelay` | `1s` | 初始重启延迟 |
| `MaxRestartDelay` | `1m` | 最大重启延迟 |
| `RestartWindow` | `5m` | 重启计数窗口期 |
| `MaxRestartsInWindow` | `5` | 窗口期最大重启次数 |
| `BackoffMultiplier` | `2.0` | 退避倍数 |

可用选项：

```go
func WithAutoStart(autoStart bool) Option
func WithRestartPolicy(policy RestartPolicy) Option
func WithMaxRestarts(n int) Option
func WithRestartDelay(d time.Duration) Option
func WithBackoff(maxDelay time.Duration, multiplier float64) Option
func WithRestartWindow(window time.Duration, maxRestarts int) Option
```

示例：

```go
_ = manager.Add(
    srv,
    supervisor.WithAutoStart(true),
    supervisor.WithRestartPolicy(supervisor.RestartOnFailure),
    supervisor.WithMaxRestarts(3),
    supervisor.WithRestartDelay(2*time.Second),
    supervisor.WithBackoff(30*time.Second, 1.5),
    supervisor.WithRestartWindow(2*time.Minute, 5),
)
```

## 错误与重启控制

- **不重启**：服务返回 `NoRestartErr(err)` 表示不触发自动重启。
- **致命错误**：使用 `AsFatalErr(err, status)` 标记为致命错误，服务会被标记为失败并停止重启。

示例：

```go
if err != nil {
    return supervisor.NoRestartErr(err)
}
```

## 调试 UI / API

Supervisor 提供调试 UI 与 API，注册入口：

```go
import "github.com/pubgo/lava/v2/core/supervisor/debug"

debug.Register(manager)
```

注册后，调试服务会挂载在调试路由下：

- `GET /supervisor/`：调试 UI
- `GET /supervisor/api/services`：服务列表
- `GET /supervisor/api/service/:name`：服务详情
- `POST /supervisor/api/service/:name/start`：启动服务
- `POST /supervisor/api/service/:name/stop`：停止服务
- `POST /supervisor/api/service/:name/restart`：重启服务
- `POST /supervisor/api/service/:name/reset`：重置失败状态与计数
- `POST /supervisor/api/services/restart`：重启所有服务

> 实际对外路径取决于调试服务器的挂载前缀与监听配置。

## 注意事项

1. `Add` 仅注册服务，不会立刻运行；需要通过 `Run/Serve` 启动管理器，或在管理器已启动时新增服务。
2. `WithAutoStart(false)` 注册的服务不会随 `Run/Serve` 自动启动，可通过 `StartService` 手动启动。
3. 建议为关键服务设置合理的 `RestartPolicy` 与窗口期上限，避免频繁崩溃导致资源消耗过高。
