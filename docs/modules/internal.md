# Internal 模块（`internal/*`）

`internal/*` 存放框架内部实现细节与示例，默认不作为对外稳定 API。

## 子目录概览

| 目录                   | 职责                                               |
| ---------------------- | -------------------------------------------------- |
| `internal/configs`     | 内部配置结构与默认值                               |
| `internal/consts`      | 内部常量                                           |
| `internal/logutil`     | 日志辅助函数                                       |
| `internal/middlewares` | 内建中间件实现                                     |
| `internal/examples`    | 示例工程（scheduler/tunnel/grpcweb/fileserver 等） |

## 使用建议

- 业务项目避免直接依赖 `internal/*`。
- 若内部能力被多处复用并趋于稳定，考虑上移到 `pkg/*` 或 `core/*`。
- 文档示例可引用 `internal/examples/*`，但应说明其“示例性质”。
