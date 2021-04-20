package clientinterceptors

import (
	"github.com/pubgo/lug/client/grpcc/clientinterceptors/hystrix"
	"google.golang.org/grpc"

)

func BreakerUnary(opts ...hystrix.Option) grpc.UnaryClientInterceptor {
	return hystrix.UnaryClientInterceptor(opts...)
}

func BreakerStream(opts ...hystrix.Option) grpc.StreamClientInterceptor {
	return hystrix.StreamClientInterceptor(opts...)
}
