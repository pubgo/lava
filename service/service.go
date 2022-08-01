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

type Options struct {
	Id        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Port      int               `json:"port,omitempty"`
	Addr      string            `json:"addr,omitempty"`
	Advertise string            `json:"advertise,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
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
