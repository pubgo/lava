package runmode

import (
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/lava/core/flags"
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
