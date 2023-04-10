package registry

import "errors"

// ErrWatcherStopped Watcher stopped error when watcher is stopped
var (
	ErrWatcherStopped = errors.New("watcher stopped")
	ErrNotFound       = errors.New("not found")
)
