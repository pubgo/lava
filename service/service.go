package service

import (
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type Init interface {
	Init()
}

type Close interface {
	Close()
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
	Providers(provider ...interface{})
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
}
