package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcvalidate "github.com/grpc-ecosystem/go-grpc-middleware/validator"
)

// DefaultUnaryMiddleware is a recommended set of middleware that should each gracefully no-op if the middleware is not
// applicable.
var DefaultUnaryMiddleware = []grpc.UnaryServerInterceptor{grpcvalidate.UnaryServerInterceptor()}

// WrapMethods wraps all non-streaming endpoints with the given list of interceptors.
// It returns a copy of the ServiceDesc with the new wrapped methods.
func WrapMethods(svcDesc grpc.ServiceDesc, interceptors ...grpc.UnaryServerInterceptor) (wrapped *grpc.ServiceDesc) {
	chain := grpcmw.ChainUnaryServer(interceptors...)
	for i, m := range svcDesc.Methods {
		handler := m.Handler
		wrapped := grpc.MethodDesc{
			MethodName: m.MethodName,
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				if interceptor == nil {
					interceptor = NoopUnaryInterceptor
				}
				return handler(srv, ctx, dec, grpcmw.ChainUnaryServer(interceptor, chain))
			},
		}
		svcDesc.Methods[i] = wrapped
	}
	return &svcDesc
}

// WrapStreams wraps all streaming endpoints with the given list of interceptors.
// It returns a copy of the ServiceDesc with the new wrapped methods.
func WrapStreams(svcDesc grpc.ServiceDesc, interceptors ...grpc.StreamServerInterceptor) (wrapped *grpc.ServiceDesc) {
	chain := grpcmw.ChainStreamServer(interceptors...)
	for i, s := range svcDesc.Streams {
		handler := s.Handler
		info := &grpc.StreamServerInfo{
			FullMethod:     fmt.Sprintf("/%s/%s", svcDesc.ServiceName, s.StreamName),
			IsClientStream: s.ClientStreams,
			IsServerStream: s.ServerStreams,
		}
		wrapped := grpc.StreamDesc{
			StreamName:    s.StreamName,
			ClientStreams: s.ClientStreams,
			ServerStreams: s.ServerStreams,
			Handler: func(srv interface{}, stream grpc.ServerStream) error {
				return chain(srv, stream, info, handler)
			},
		}
		svcDesc.Streams[i] = wrapped
	}
	return &svcDesc
}

// NoopUnaryInterceptor is a gRPC middleware that does not do anything.
func NoopUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return handler(ctx, req)
}

// WithCustomCerts is a dial option for requiring TLS from a specified path.
// If the path is a directory, all certs are loaded. If it is an individual
// file only the directly specified cert is loaded.
//
// This function panics if custom certificate pool cannot be instantiated.
func WithCustomCerts(certPath string, insecureSkipVerify bool) grpc.DialOption {
	certPool, err := customCertPool(certPath)
	if err != nil {
		panic(err)
	}

	return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: insecureSkipVerify,
	}))
}

// WithSystemCerts is a dial option for requiring TLS with the system
// certificate pool.
//
// This function panics if the system pool cannot be loaded.
func WithSystemCerts(insecureSkipVerify bool) grpc.DialOption {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: insecureSkipVerify,
	}))
}
