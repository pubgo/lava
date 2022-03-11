package config

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

func DefaultFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Destination: &CfgPath,
			Aliases:     types.StrList{"c"},
			Usage:       "config path",
			Value:       CfgPath,
		},
		&cli.StringFlag{
			Name:        "addr",
			Destination: &runtime.Addr,
			Aliases:     types.StrList{"a"},
			Usage:       "server(http|grpc|ws|...) address",
			Value:       runtime.Addr,
		},
		&cli.BoolFlag{
			Name:        "trace",
			Destination: &runtime.Trace,
			Aliases:     types.StrList{"t"},
			Usage:       "enable trace",
			Value:       runtime.Trace,
			EnvVars:     env.KeyOf("trace", "trace-log", "tracelog"),
		},
		// 运行环境
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runtime.Mode,
			Aliases:     types.StrList{"m"},
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runtime.Mode,
			EnvVars:     env.KeyOf("lava-mode", "lava.mode"),
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runtime.Level,
			Aliases:     types.StrList{"l"},
			Usage:       "log level(debug|info|warn|error|panic|fatal)",
			Value:       runtime.Level,
		},
	}
}
