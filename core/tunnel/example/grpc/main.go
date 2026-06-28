// Demo: gRPC over tunnel —— 经网关 gRPC 代理调用后端 gRPC 服务。
//
// 演示：
//   - 后端注册 grpc 端点
//   - tunnel.GRPCContextDialer 封装路由行（TUNNEL <service>\n）与可选 token
//   - 标准 grpc.NewClient + health check 调用
//
// 运行：
//
//	go run ./core/tunnel/example/grpc
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
)

const (
	gatewayAddr = "127.0.0.1:17300"
	grpcPort    = 19300             // 网关对外 gRPC 代理端口
	grpcBackend = "127.0.0.1:19301" // 本地 gRPC 服务
	serviceName = "grpc-svc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1) 本地 gRPC health 服务
	stopBackend := startGRPCBackend()
	defer stopBackend()

	// 2) 网关（开放 gRPC 代理端口）
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gatewayAddr,
		Transport:  tunnel.TransportYamux,
		GRPCPort:   grpcPort,
	})
	if err := gw.Start(ctx); err != nil {
		log.Fatalf("gateway: %v", err)
	}
	defer gw.Stop(context.Background())

	// 3) Agent：注册 grpc 端点
	agent, err := tunnelagent.Standalone(ctx, tunnelagent.StandaloneOptions{
		GatewayAddr: gatewayAddr,
		ServiceName: serviceName,
		Endpoints:   []tunnel.EndpointConfig{{Type: "grpc", LocalAddr: grpcBackend}},
	})
	if err != nil {
		log.Fatalf("agent: %v", err)
	}
	defer agent.Stop(context.Background())
	time.Sleep(500 * time.Millisecond)

	// 4) gRPC 客户端经网关代理拨号（addr 即 service 名）
	cc, err := grpc.NewClient("passthrough:///"+serviceName,
		grpc.WithContextDialer(tunnel.GRPCContextDialer(tunnel.GRPCDialOptions{
			GatewayAddr: fmt.Sprintf("127.0.0.1:%d", grpcPort),
			// Token: "secret", // 网关启用 auth 时填写
		})),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("grpc client: %v", err)
	}
	defer cc.Close()

	resp, err := grpc_health_v1.NewHealthClient(cc).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("health check: %v", err)
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("gRPC over tunnel OK, health status: %s\n", resp.GetStatus())
	fmt.Println("----------------------------------------")
}

func startGRPCBackend() func() {
	lis, err := net.Listen("tcp", grpcBackend)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	srv := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	go func() { _ = srv.Serve(lis) }()
	return func() {
		srv.Stop()
		_ = lis.Close()
	}
}
