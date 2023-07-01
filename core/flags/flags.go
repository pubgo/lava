package flags

import (
	"github.com/urfave/cli/v3"
)

var flags []cli.Flag

func Register(flag cli.Flag) {
	flags = append(flags, flag)
}

func GetFlags() []cli.Flag { return flags }
