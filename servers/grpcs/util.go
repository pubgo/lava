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

func parsePath(path string) (string, string, string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", ""
	}

	method := parts[len(parts)-1]
	pkgService := parts[len(parts)-2]
	prefix := strings.Join(parts[0:len(parts)-2], "/")
	return prefix, pkgService, method
}
