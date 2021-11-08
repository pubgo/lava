package clix

import (
	"strings"

	"github.com/urfave/cli/v2"
)

func ExampleFmt(data ...string) string {
	var str = ""
	for i := range data {
		str += "  " + data[i] + "\n"
	}
	return "  " + strings.TrimSpace(str)
}

type Flags []cli.Flag
