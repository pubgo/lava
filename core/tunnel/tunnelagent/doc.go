// Package tunnelagent 提供隧道代理客户端的便捷封装
//
// 这个包是 tunnel.Agent 的包装器，让用户可以更方便地创建和管理隧道代理客户端。
//
// # 基本用法
//
// 创建一个连接到 Gateway 的 Agent:
//
//	agent := tunnelagent.New(&tunnelagent.Config{
//		GatewayAddr: "gateway.example.com:9000",
//		Transport:   "yamux",  // 或 "quic", "kcp"
//		ServiceName: "my-service",
//		ServiceVersion: "v1.0.0",
//		Endpoints: []tunnelagent.EndpointConfig{
//			{Type: "http", LocalAddr: ":8080"},
//			{Type: "grpc", LocalAddr: ":9090"},
//			{Type: "debug", LocalAddr: ":6060"},
//		},
//	})
//
//	ctx := context.Background()
//	if err := agent.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer agent.Stop(ctx)
//
// # 动态服务注册
//
// 可以在运行时动态注册和注销服务:
//
//	// 注册新服务
//	err := agent.Register(ctx, &tunnelagent.ServiceInfo{
//		Name:    "new-service",
//		Version: "v1.0.0",
//		Endpoints: []tunnelagent.Endpoint{
//			{Type: "http", Address: ":8081"},
//		},
//	})
//
//	// 注销服务
//	err = agent.Deregister(ctx, "new-service")
//
// # 传输协议
//
// 支持多种传输协议:
//
//   - yamux: 基于 TCP 的多路复用（默认）
//   - quic: 基于 UDP 的 QUIC 协议
//   - kcp: 基于 UDP 的 KCP 协议
//
// 选择传输协议时需要考虑:
//
//   - yamux: 适合大多数场景，稳定可靠
//   - quic: 适合高延迟或丢包网络，支持 0-RTT
//   - kcp: 适合对延迟敏感但可接受较高带宽的场景
//
// # TLS 配置
//
// 通过 TransportOptions 配置 TLS:
//
//	agent := tunnelagent.New(&tunnelagent.Config{
//		GatewayAddr: "gateway.example.com:9000",
//		Transport:   "yamux",
//		TransportOptions: &tunnelagent.TransportOptions{
//			EnableTLS: true,
//			CertFile:  "/path/to/client.crt",
//			KeyFile:   "/path/to/client.key",
//			CAFile:    "/path/to/ca.crt",
//		},
//		ServiceName: "my-service",
//	})
//
// # 状态监控
//
// 检查 Agent 状态:
//
//	status := agent.Status()
//	switch status {
//	case tunnelagent.StatusConnected:
//		fmt.Println("已连接到 Gateway")
//	case tunnelagent.StatusReconnecting:
//		fmt.Println("正在重连...")
//	case tunnelagent.StatusDisconnected:
//		fmt.Println("未连接")
//	}
//
//	info := agent.Info()
//	fmt.Printf("Gateway: %s, Service: %s\n", info.GatewayAddr, info.ServiceName)
package tunnelagent
