package config

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/urfave/cli/v2"
)

func init() {
	defer recovery.Exit()

	flags.Register(&cli.StringFlag{
		Name:        "home",
		Destination: &CfgPath,
		Usage:       "config home dir, [configs]",
		EnvVars:     typex.StrOf(consts.EnvHome),
	})
}
