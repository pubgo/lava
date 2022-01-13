package mux

import "strings"

func DebugPrefix(names ...string) string {
	var ns []string
	for i := range names {
		ns = append(ns, strings.Trim(names[i], "/"))
	}
	return "/debug/" + strings.Join(ns, "/")
}
