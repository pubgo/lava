package registry

import "errors"

var (
	// ErrWatcherStopped 表示 Watcher 已停止。
	ErrWatcherStopped = errors.New("watcher stopped")
	// ErrNotFound 表示请求的服务不存在。
	ErrNotFound = errors.New("not found")
)
