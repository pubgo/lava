package flags

import (
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/typex"
	"github.com/urfave/cli/v2"
)

func init() {
	Register(&cli.IntFlag{
		Name:        "http-port",
		Usage:       "service http port",
		Value:       running.HttpPort,
		Destination: &running.HttpPort,
		EnvVars:     typex.StrOf(env.Key("server_http_port")),
	})
	Register(&cli.IntFlag{
		Name:        "grpc-port",
		Usage:       "service grpc port",
		Value:       running.GrpcPort,
		Destination: &running.GrpcPort,
		EnvVars:     typex.StrOf(env.Key("server_grpc_port")),
	})
	Register(&cli.BoolFlag{
		Name:        "debug",
		Usage:       "enable debug mode",
		Value:       running.IsDebug,
		Destination: &running.IsDebug,
		EnvVars:     typex.StrOf(env.Key("debug"), env.Key("enable_debug")),
	})
	Register(&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "config path",
		Value:   config.GetConfigPath(),
		EnvVars: typex.StrOf(env.Key("config_path")),
		Action: func(context *cli.Context, s string) error {
			config.SetConfigPath(s)
			return nil
		},
	})
}
