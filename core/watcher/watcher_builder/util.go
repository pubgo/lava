package watcher_builder

import (
	"strings"

	"github.com/pubgo/lava/runtime"
)

func trimProject(key string) string {
	return strings.Trim(strings.TrimPrefix(key, runtime.Name()), ".")
}
