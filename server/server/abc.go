package server

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"io"
	"net"
)

type Service interface {
	AfterStop(func())
	BeforeStop(func())
	AfterStart(func())
	BeforeStart(func())
	Middleware(middleware types.Middleware)
	Description(description ...string)
	Flags(flags cli.Flag)
	RegisterMatcher(priority int64, matches ...func(io.Reader) bool) func() net.Listener
	Commands(commands *cli.Command)
	InnerGrpcClientConn() grpc.ClientConnInterface
	ServiceDesc() ServiceDesc
	Plugin(plugin plugin.Plugin)
	Options() Options
	start() error
	stop() error
}

type ServiceDesc struct {
	grpc.ServiceDesc
	Handler       interface{}
	GrpcClientFn  interface{}
	GrpcGatewayFn interface{}
}

type Handler interface {
	Init() func() error
	Flags() types.Flags
}

type Options struct {
	Id       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Version  string            `json:"version,omitempty"`
	Port     int               `json:"port,omitempty"`
	Address  string            `json:"address,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
