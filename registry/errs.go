package registry

import "errors"

// ErrWatcherStopped Watcher stopped error when watcher is stopped
var ErrWatcherStopped = errors.New("[registry] watcher stopped")
var ErrNotFound = errors.New("[registry] not found")
