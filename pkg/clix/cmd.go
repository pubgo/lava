package clix

import (
	"strings"
)

func ExampleFmt(data ...string) string {
	var str = ""
	for i := range data {
		str += "  " + data[i] + "\n"
	}
	return "  " + strings.TrimSpace(str)
}
