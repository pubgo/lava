package lava

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

type ProxyCfg struct {
	// Name service name
	Name string `yaml:"name"`

	// Addr service address
	Addr string `yaml:"addr"`

	// Resolver service resolver, default direct
	Resolver string `yaml:"resolver"`
}

type GrpcProxy interface {
	GrpcRouter
	Proxy() ProxyCfg
}

type GrpcRouter interface {
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
}

type HttpRouter interface {
	Middlewares() []Middleware
	Router(router fiber.Router)
	Prefix() string
	// Annotation() []Annotation
}
