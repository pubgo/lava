package config

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

func DefaultFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Destination: &CfgPath,
			Aliases:     typex.StrOf("c"),
			Usage:       "config path",
			Value:       CfgPath,
		},
		&cli.StringFlag{
			Name:        "addr",
			Destination: &runenv.Addr,
			Aliases:     typex.StrOf("a"),
			Usage:       "server(http|grpc|ws|...) address",
			Value:       runenv.Addr,
		},
		&cli.BoolFlag{
			Name:        "trace",
			Destination: &runenv.Trace,
			Aliases:     typex.StrOf("t"),
			Usage:       "enable trace",
			Value:       runenv.Trace,
			EnvVars:     types.EnvOf("trace", "trace-log", "tracelog"),
		},
		// 运行环境
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runenv.Mode,
			Aliases:     typex.StrOf("m"),
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runenv.Mode,
			EnvVars:     types.EnvOf("lava-run-mode", "lava-run-env"),
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runenv.Level,
			Aliases:     typex.StrOf("l"),
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
