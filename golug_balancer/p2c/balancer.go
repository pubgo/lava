package p2c

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Name = "p2c"

func init() {
	balancer.Register(base.NewBalancerBuilderWithConfig(Name, &builder{}, base.Config{HealthCheck: true}))
}
