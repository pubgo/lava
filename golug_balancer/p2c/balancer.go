package p2c

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Name = "p2c"

func RegisterBalancer() {
	balancer.Register(base.NewBalancerBuilderWithConfig(Name, &p2cBalancer{}, base.Config{HealthCheck: true}))
}

type baseBalancer struct{}
