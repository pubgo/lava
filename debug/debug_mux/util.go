package debug_mux

import (
	"strings"

	"github.com/pubgo/lava/types"
)

func DebugPrefix(names ...string) string {
	var ns = types.StrList{"debug"}
	for i := range names {
		var p = strings.TrimSpace(strings.Trim(names[i], "/"))
		if p == "" {
			continue
		}

		ns = append(ns, p)
	}
	return strings.TrimSpace("/" + strings.Join(ns, "/"))
}
