package flags

import (
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/typex"
	"github.com/urfave/cli/v3"
)

func init() {
	Register(&cli.IntFlag{
		Name:        "http-port",
		Usage:       "service http port",
		Value:       running.HttpPort,
		Destination: &running.HttpPort,
		EnvVars:     typex.StrOf(env.Key("http_port")),
	})
	Register(&cli.IntFlag{
		Name:        "grpc-port",
		Usage:       "service grpc port",
		Value:       running.GrpcPort,
		Destination: &running.GrpcPort,
		EnvVars:     typex.StrOf(env.Key("grpc_port")),
	})
}
