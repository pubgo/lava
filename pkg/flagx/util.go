package flagx

import "github.com/urfave/cli/v2"

var _ cli.Generic = (*Generic)(nil)

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
