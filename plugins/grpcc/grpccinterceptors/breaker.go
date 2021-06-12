package grpccinterceptors

import (
	hystrix2 "github.com/pubgo/lug/plugins/grpcc/grpccinterceptors/hystrix"
	"google.golang.org/grpc"
)

func BreakerUnary(opts ...hystrix2.Option) grpc.UnaryClientInterceptor {
	return hystrix2.UnaryClientInterceptor(opts...)
}

func BreakerStream(opts ...hystrix2.Option) grpc.StreamClientInterceptor {
	return hystrix2.StreamClientInterceptor(opts...)
}
