package lava

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

type GrpcProxyCfg struct {
	// Name grpc service name
	Name string `yaml:"name"`

	// Addr grpc service address, dns://auth:8080, auth:8080
	Addr string `yaml:"addr"`

	// Resolver service resolver[direct, k8s, dns, etc...], default direct
	Resolver string `yaml:"resolver"`
}

type GrpcProxy interface {
	GrpcRouter
	Proxy() GrpcProxyCfg
}

type GrpcHttpRouter interface {
	GrpcRouter
	Router(router fiber.Router)
	Prefix() string
}

type GrpcRouter interface {
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
}

type HttpRouter interface {
	Middlewares() []Middleware
	Router(router fiber.Router)

	// Prefix router prefix, required
	Prefix() string
}
