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

type Command interface {
	Command() *cli.Command
	Start() error
	Stop() error
}

type Service interface {
	grpc.ServiceRegistrar
	Command
	Options() Options
	Provider(provider interface{})
	SubCmd(cmd *cli.Command)
}

type Web interface {
	Command
	Options() Options
	Provider(provider interface{})
	SubCmd(cmd *cli.Command)
}
