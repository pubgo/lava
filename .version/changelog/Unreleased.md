## [Unreleased]

<!-- 在此记录尚未发版的变更；发版时移至 CHANGELOG.md 并标注版本号。 -->

### Changed

- **metrics**: RPC 指标名统一为 `lava_rpc_total` / `lava_rpc_failed_total` / `lava_rpc_handling_seconds`，客户端与服务端共用一组，用 `side`/`kind`/`service`/`method`/`stream`/`proto`（失败另有 `code`）区分；`gateway_server_*` 已删除，抓取端与看板需同步改。
