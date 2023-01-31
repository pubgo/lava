package config

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/flags"
)

func init() {
	flags.Register(&cli.StringFlag{
		Name:        "home",
		Destination: &CfgPath,
		Usage:       "config home dir, [configs]",
		EnvVars:     typex.StrOf(env.Key(consts.EnvHome)),
	})

	flags.Register(&cli.StringFlag{
		Name:  "config",
		Usage: "config file name",
		Value: FileName + "." + FileType,
	})
}
