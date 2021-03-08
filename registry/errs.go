package registry

import "errors"

// Not found error when GetService is called
var ErrNotFound = errors.New("not found")

// Watcher stopped error when watcher is stopped
var ErrWatcherStopped = errors.New("watcher stopped")
