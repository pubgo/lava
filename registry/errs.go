package registry

import (
	"github.com/pubgo/xerror"
)

var Err = xerror.New(Name)

// ErrWatcherStopped Watcher stopped error when watcher is stopped
var ErrWatcherStopped = Err.New("watcher stopped")
var ErrNotFound = Err.New("not found")
