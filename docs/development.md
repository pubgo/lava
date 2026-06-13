# Lava 开发指南

## 1. 开发环境

- Go `1.25+`
- Task
- `golangci-lint`（执行 `task lint` 需要）

## 2. 仓库初始化

```bash
git clone https://github.com/pubgo/lava.git
cd lava
go mod tidy
```

## 3. 常用开发命令（以 `Taskfile.yml` 为准）

| 命令              | 说明                                   |
| ----------------- | -------------------------------------- |
| `task info`       | 输出版本与分支信息                     |
| `task tools`      | 安装常用开发工具                       |
| `task generate`   | 执行 `go generate ./internal/...`      |
| `task proto:fmt`  | proto 格式化                           |
| `task proto:lint` | proto lint                             |
| `task proto:gen`  | proto 依赖拉取并生成代码               |
| `task vet`        | `go vet ./...`                         |
| `task test`       | `go test -short -race -v ./... -cover` |
| `task lint`       | `golangci-lint run --verbose ./...`    |

## 4. 推荐开发流程

1. 新建分支开发
2. 小步提交（建议 `feat/fix/docs/refactor/test/chore` 前缀）
3. 每次提交前执行：

```bash
task test
task lint
```

4. 涉及 proto 变更时执行：

```bash
task proto:fmt
task proto:lint
task proto:gen
```

说明：`task proto:gen` 现在会自动安装本地 `zrpc` 插件：

- `go install ./tools/protoc-gen-zrpc-go`

如果你新增了 `zrpc` 相关 `.proto`，无需手动处理插件路径，直接执行任务即可。

## 5. 命令入口说明

仓库当前存在两种入口：

- 根入口：`main.go`（工具型命令）
- DI 入口：`core/lavabuilder.Run`（服务型命令装配）

开发文档中的命令示例默认以根入口为准。

## 6. 文档维护规范

- 命令文档必须对齐 `main.go` 实际注册命令。
- 接口文档必须对齐 `lava/*.go` 与 `core/supervisor/types.go`。
- 涉及流程图更新时，需同步标注对应实现路径。

## 7. 常见问题

### Q1: `task build` / `task clean` 为什么不存在？

当前根 `Taskfile.yml` 未定义这些任务，请以 `task -a` 输出为准。

### Q2: 为什么 `cmds/` 有些命令跑不出来？

因为它们未接入 `main.go`。详见 `docs/modules/cmds.md`。

### Q3: 我该先看哪份文档？

建议顺序：

1. `docs/architecture-v2.md`
2. `docs/design-v2.md`
3. `docs/zrpc.md`（如果你在看 zrpc/NATS RPC）
4. `docs/modules/README.md`
5. `docs/lava-command.md`
