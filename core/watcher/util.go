package watcher

import (
	"strings"
)

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
