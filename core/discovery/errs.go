package discovery

import "errors"

var (
	// ErrWatcherStopped 表示 Watcher 已停止，Next 不应再被调用。
	ErrWatcherStopped = errors.New("watcher stopped")
	// ErrTimeout 表示等待服务变更超时。
	ErrTimeout = errors.New("watcher timeout")
)
