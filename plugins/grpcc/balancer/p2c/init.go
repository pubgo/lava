package p2c

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Name = "p2c"

func init() {
	// 注册balancer到grpc balancer管理中心, 后期如果要使用的话, 需要指定本balancer的名字[Name]
	balancer.Register(base.NewBalancerBuilder(Name, &p2cBalancer{}, base.Config{HealthCheck: true}))
}
