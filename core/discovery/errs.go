package discovery

import "errors"

// ErrWatcherStopped Watcher stopped error when watcher is stopped
var ErrWatcherStopped = errors.New("watcher stopped")
var ErrNotFound = errors.New("not found")
