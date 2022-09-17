package runmode

import (
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/urfave/cli/v2"
)

func init() {
	flags.Register(&cli.IntFlag{
		Name:        "http-port",
		Usage:       "service http port",
		Value:       HttpPort,
		Destination: &HttpPort,
		EnvVars:     typex.StrOf(env.Key("http_port")),
	})
	flags.Register(&cli.IntFlag{
		Name:        "grpc-port",
		Usage:       "service grpc port",
		Value:       GrpcPort,
		Destination: &GrpcPort,
		EnvVars:     typex.StrOf(env.Key("grpc_port")),
	})
}
