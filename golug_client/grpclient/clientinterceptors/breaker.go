package clientinterceptors

import (
	"github.com/pubgo/golug/golug_client/grpclient/clientinterceptors/grpc_hystrix"
	"google.golang.org/grpc"

)

func BreakerUnary(opts ...grpc_hystrix.Option) grpc.UnaryClientInterceptor {
	return grpc_hystrix.UnaryClientInterceptor(opts...)
}

func BreakerStream(opts ...grpc_hystrix.Option) grpc.StreamClientInterceptor {
	return grpc_hystrix.StreamClientInterceptor(opts...)
}
