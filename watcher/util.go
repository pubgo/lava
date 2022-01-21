package watcher

import (
	"strings"

	"github.com/pubgo/lava/runtime"
)

func trimProject(key string) string {
	return strings.Trim(strings.TrimPrefix(key, runtime.Name()), ".")
}

// KeyToDot /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix ...string) string {
	var p string
	if len(prefix) > 0 {
		p = strings.Join(prefix, ".")
	}

	p = strings.ReplaceAll(strings.ReplaceAll(p, "/", "."), "..", ".")
	p = strings.Trim(p, ".")

	return p
}
