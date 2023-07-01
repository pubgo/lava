package grpcs

import "strings"

// serviceFromMethod returns the service
// /service.Foo/Bar => service.Foo
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	return strings.Split(strings.Trim(m, "/"), "/")[0]
}
