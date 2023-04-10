package cmdutil

import (
	"fmt"
	"strings"
)

func ExampleFmt(data ...string) string {
	str := ""
	for i := range data {
		str += "  " + data[i] + "\n"
	}
	return "  " + strings.TrimSpace(str)
}

func UsageDesc(format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args...)
	return strings.ToUpper(s[0:1]) + s[1:]
}
