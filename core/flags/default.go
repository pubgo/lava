package flags

import (
	"context"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/running"
	"github.com/urfave/cli/v3"
)

func init() {
	const httpPort = "server_http_port"
	const grpcPort = "server_grpc_port"
	const conf = "config_path"
	env.GetIntVal(&running.HttpPort, httpPort)
	env.GetIntVal(&running.GrpcPort, grpcPort)

	Register(&cli.IntFlag{
		Name:    "http-port",
		Usage:   "service http port",
		Local:   true,
		Value:   running.HttpPort,
		Sources: cli.EnvVars(env.Key(httpPort)),
		Action: func(ctx context.Context, command *cli.Command, i int) error {
			running.HttpPort = i
			return nil
		},
	})

	Register(&cli.IntFlag{
		Name:    "grpc-port",
		Usage:   "service grpc port",
		Local:   true,
		Value:   running.GrpcPort,
		Sources: cli.EnvVars(env.Key(grpcPort)),
		Action: func(ctx context.Context, command *cli.Command, i int) error {
			running.GrpcPort = i
			return nil
		},
	})

	Register(&cli.BoolFlag{
		Name:        "debug",
		Usage:       "enable debug mode",
		Local:       true,
		Value:       running.IsDebug,
		Destination: &running.IsDebug,
		Sources:     cli.EnvVars(env.Key("debug"), env.Key("enable_debug")),
	})

	Register(&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "config path",
		Value:   config.GetConfigPath(),
		Local:   true,
		Sources: cli.EnvVars(env.Key(conf)),
		Action: func(ctx context.Context, command *cli.Command, s string) error {
			config.SetConfigPath(s)
			return nil
		},
	})
}
