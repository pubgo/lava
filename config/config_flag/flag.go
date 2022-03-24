package config_flag

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runtime"
)

func DefaultFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Destination: &config.CfgPath,
			Aliases:     typex.StrOf("c"),
			Usage:       "config path",
			Value:       config.CfgPath,
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
			EnvVars:     env.KeyOf("trace", "trace-log", "tracelog"),
		},
		// 运行环境
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runtime.Mode,
			Aliases:     typex.StrOf("m"),
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runtime.Mode,
			EnvVars:     env.KeyOf("lava-mode", "lava.mode"),
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runtime.Level,
			Aliases:     typex.StrOf("l"),
			Usage:       "log level(debug|info|warn|error|panic|fatal)",
			Value:       runtime.Level,
		},
	}
}
