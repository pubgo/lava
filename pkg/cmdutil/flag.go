package cmdutil

import (
	"flag"
	"os"
	"strings"
)

var _ flag.Value = (*Generic)(nil)

type Generic struct {
	Value       string
	Destination func(val string) error
}

func (f Generic) Set(value string) error {
	return f.Destination(value)
}

func (f Generic) String() string {
	return f.Value
}

func IsHelp() bool {
	var arg = strings.TrimSpace(os.Args[len(os.Args)-1])
	return arg == "--help" || arg == "-h"
}
