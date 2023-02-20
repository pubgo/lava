package flags

import (
	"github.com/pubgo/funk/env"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/typex"
	"github.com/urfave/cli/v3"
)

func init() {
	Register(&cli.IntFlag{
		Name:        "http-port",
		Usage:       "service http port",
		Value:       runmode.HttpPort,
		Destination: &runmode.HttpPort,
		EnvVars:     typex.StrOf(env.Key("http_port")),
	})
	Register(&cli.IntFlag{
		Name:        "grpc-port",
		Usage:       "service grpc port",
		Value:       runmode.GrpcPort,
		Destination: &runmode.GrpcPort,
		EnvVars:     typex.StrOf(env.Key("grpc_port")),
	})
}
