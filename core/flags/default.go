package flags

import (
	"context"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/running"
	"github.com/urfave/cli/v3"
)

func init() {
	Register(&cli.IntFlag{
		Name:       "http-port",
		Usage:      "service http port",
		Persistent: true,
		Value:      int64(running.HttpPort),
		Sources:    cli.EnvVars(env.Key("server_http_port")),
		Action: func(ctx context.Context, command *cli.Command, i int64) error {
			running.HttpPort = int(i)
			return nil
		},
	})
	Register(&cli.IntFlag{
		Name:       "grpc-port",
		Usage:      "service grpc port",
		Persistent: true,
		Value:      int64(running.GrpcPort),
		Sources:    cli.EnvVars(env.Key("server_grpc_port")),
		Action: func(ctx context.Context, command *cli.Command, i int64) error {
			running.GrpcPort = int(i)
			return nil
		},
	})
	Register(&cli.BoolFlag{
		Name:        "debug",
		Usage:       "enable debug mode",
		Persistent:  true,
		Value:       running.IsDebug,
		Destination: &running.IsDebug,
		Sources:     cli.EnvVars(env.Key("debug"), env.Key("enable_debug")),
	})
	Register(&cli.StringFlag{
		Name:       "config",
		Aliases:    []string{"c"},
		Usage:      "config path",
		Value:      config.GetConfigPath(),
		Persistent: true,
		Sources:    cli.EnvVars(env.Key("config_path")),
		Action: func(ctx context.Context, command *cli.Command, s string) error {
			config.SetConfigPath(s)
			return nil
		},
	})
}
