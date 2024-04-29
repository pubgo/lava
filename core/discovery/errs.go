package discovery

import "errors"

var (
	ErrWatcherStopped = errors.New("err watcher stopped")
	ErrTimeout        = errors.New("err watcher timeout")
)
