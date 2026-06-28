package tunnel_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func grpcHealthDial(_ context.Context, gatewayGRPCAddr, serviceName, token string) (*grpc.ClientConn, grpc_health_v1.HealthClient, error) {
	cc, err := grpc.NewClient("passthrough:///"+serviceName,
		grpc.WithContextDialer(tunnel.GRPCContextDialer(tunnel.GRPCDialOptions{
			GatewayAddr: gatewayGRPCAddr,
			Token:       token,
		})),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}
	return cc, grpc_health_v1.NewHealthClient(cc), nil
}

func startGRPCHealthServer(t *testing.T, addr string) (*grpc.Server, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	go func() { _ = srv.Serve(lis) }()
	return srv, func() {
		srv.Stop()
		_ = lis.Close()
	}
}
