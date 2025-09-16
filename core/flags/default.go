package flags

import (
	"context"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/running"
	"github.com/urfave/cli/v3"
)

func init() {
	var httpPortEnvs = []string{env.Key("server_http_port"), env.Key("service_http_port")}
	var grpcPortEnvs = []string{env.Key("server_grpc_port"), env.Key("service_grpc_port")}
	const conf = "config_path"
	if port := env.GetInt(httpPortEnvs...); port > 0 {
		running.HttpPort = port
	}

	if port := env.GetInt(grpcPortEnvs...); port > 0 {
		running.GrpcPort = port
	}

	Register(&cli.IntFlag{
		Name:    "http-port",
		Usage:   "service http port",
		Local:   true,
		Value:   running.HttpPort,
		Sources: cli.EnvVars(httpPortEnvs...),
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
		Sources: cli.EnvVars(grpcPortEnvs...),
		Action: func(ctx context.Context, command *cli.Command, i int) error {
			running.GrpcPort = i
			return nil
		},
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

	Register(&cli.StringFlag{
		Name:    "env",
		Usage:   "runtime env",
		Value:   running.Env,
		Local:   true,
		Sources: cli.EnvVars(env.Key("env"), env.Key("app_env")),
		Action: func(ctx context.Context, command *cli.Command, s string) error {
			running.Env = s
			return nil
		},
	})
}
