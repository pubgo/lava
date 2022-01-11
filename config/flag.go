package config

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runenv"
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
			Destination: &runenv.Addr,
			Aliases:     types.StrList{"a"},
			Usage:       "server(http|grpc|ws|...) address",
			Value:       runenv.Addr,
		},
		&cli.BoolFlag{
			Name:        "trace",
			Destination: &runenv.Trace,
			Aliases:     types.StrList{"t"},
			Usage:       "enable trace",
			Value:       runenv.Trace,
			EnvVars:     env.KeyOf("trace", "trace-log", "tracelog"),
		},
		// 运行环境
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runenv.Mode,
			Aliases:     types.StrList{"m"},
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runenv.Mode,
			EnvVars:     env.KeyOf("lava-run-mode", "lava-run-env"),
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runenv.Level,
			Aliases:     types.StrList{"l"},
			Usage:       "log level(debug|info|warn|error|panic|fatal)",
			Value:       runenv.Level,
		},
		&cli.BoolFlag{
			Name:        "catch-sigpipe",
			Destination: &runenv.CatchSigpipe,
			Usage:       "catch and ignore SIGPIPE on stdout and stderr if specified",
			Value:       runenv.CatchSigpipe,
		},
	}
}
