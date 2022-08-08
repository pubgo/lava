package config

import (
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/typex"
)

func init() {
	defer recovery.Exit()

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
