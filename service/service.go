package service

import (
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type Init interface {
	Init() error
}

type Close interface {
	Close() error
}

type Flags interface {
	Flags() []cli.Flag
}

type Runtime interface {
	Start()
	Stop()
	Run()
}

type Service interface {
	Runtime
	Provider(provider interface{})
	RegisterGateway(register ...GatewayRegister)
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
	RegisterServer(register interface{}, impl interface{})
}
