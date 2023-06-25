package discovery

import "errors"

var ErrWatcherStopped = errors.New("err watcher stopped")
var ErrTimeout = errors.New("err watcher timeout")
