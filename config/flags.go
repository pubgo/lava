package config

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runtime"
)

func init() {
	flags.Register(&cli.StringFlag{
		Name:        "config",
		Destination: &CfgPath,
		Aliases:     typex.StrOf("c"),
		Usage:       "server config path",
	})

	flags.Register(&cli.StringFlag{
		Name:        "addr",
		Destination: &runtime.Addr,
		Aliases:     typex.StrOf("a"),
		Usage:       "server(http|grpc|ws|...) address",
		EnvVars:     []string{"app.addr"},
		Value:       runtime.Addr,
	})

	flags.Register(&cli.BoolFlag{
		Name:        "trace",
		Destination: &runtime.Trace,
		Aliases:     typex.StrOf("t"),
		Usage:       "enable trace",
		Value:       runtime.Trace,
		EnvVars:     []string{"app.trace"},
	})

	flags.Register(&cli.StringFlag{
		Name:        "level",
		Destination: &runtime.Level,
		Aliases:     typex.StrOf("l"),
		Usage:       "log level(debug|info|warn|error|panic|fatal)",
		EnvVars:     []string{"app.level"},
		Value:       runtime.Level,
	})
}
