package config_flag

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/flagx"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runtime"
)

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Destination: &config.CfgPath,
			Aliases:     typex.StrOf("c"),
			Usage:       "server config path",
		},
		&cli.StringFlag{
			Name:        "addr",
			Destination: &runtime.Addr,
			Aliases:     typex.StrOf("a"),
			Usage:       "server(http|grpc|ws|...) address",
			Value:       runtime.Addr,
		},
		&cli.BoolFlag{
			Name:        "trace",
			Destination: &runtime.Trace,
			Aliases:     typex.StrOf("t"),
			Usage:       "enable trace",
			Value:       runtime.Trace,
			EnvVars:     env.KeyOf("trace"),
		},
		// 运行环境
		&cli.GenericFlag{
			Name:    "mode",
			Aliases: typex.StrOf("m"),
			Usage:   "running mode(dev|test|stag|prod|release)",
			EnvVars: env.KeyOf("lava-mode", "lava.mode"),
			Value: flagx.Generic{
				Value: runtime.Mode.String(),
				Destination: func(val string) error {
					var i, err = strconv.Atoi(val)
					if err != nil {
						return err
					}
					runtime.Mode = runtime.RunMode(i)
					if runtime.Mode == runtime.RunModeUnknown {
						return fmt.Errorf("unknown mode, mode=%s", val)
					}
					return nil
				},
			},
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runtime.Level,
			Aliases:     typex.StrOf("l"),
			Usage:       "log level(debug|info|warn|error|panic|fatal)",
			EnvVars:     env.KeyOf("lava-level", "lava.level"),
			Value:       runtime.Level,
		},
	}
}
