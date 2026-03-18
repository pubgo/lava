# Lava 文档总览

这套文档以当前仓库实现为准，重点覆盖：

- 架构与运行流程
- 设计原则与核心抽象
- 各模块职责与入口
- CLI 使用方式

## 推荐阅读顺序

1. `architecture-v2.md`：先建立全局认知
2. `design-v2.md`：理解抽象与设计取舍
3. `modules/README.md`：按目录快速定位模块
4. `lava-command.md`：查命令参数与示例

## 文档地图

- 架构：`architecture-v2.md`
- 设计：`design-v2.md`
- 命令：`lava-command.md`
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
