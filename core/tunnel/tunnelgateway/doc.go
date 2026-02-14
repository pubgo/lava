// Package tunnelgateway 提供隧道代理网关的便捷封装
//
// 这个包是 tunnel.Gateway 的包装器，让用户可以更方便地创建和管理隧道代理网关。
//
// # 基本用法
//
// 创建一个接受 Agent 连接的 Gateway:
//
//	gw := tunnelgateway.New(&tunnelgateway.Config{
//		ListenAddr: ":9000",
//		Transport:  "yamux",  // 或 "quic", "kcp"
//		HTTPPort:   8080,     // 对外暴露的 HTTP 端口
//		GRPCPort:   9090,     // 对外暴露的 gRPC 端口
//		DebugPort:  6060,     // 对外暴露的 Debug 端口
//	})
//
//	ctx := context.Background()
//	if err := gw.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer gw.Stop(ctx)

// # 启动与验证
//
// 1) 启动 Gateway（监听 Agent 连接端口 + 对外 HTTP 代理端口）：
//
//	gw := tunnelgateway.New(&tunnelgateway.Config{
//		ListenAddr: ":7000", // Agent 连接端口
//		Transport:  "yamux",
//		HTTPPort:   8080,    // 对外 HTTP 代理端口
//	})
//
// 2) 启动 Agent 并注册服务后，访问路由：
//
//	GET http://gateway:8080/{service_name}/{path}
//
// 例如服务名为 demo-svc，后端路由 /hello：
//
//	curl http://gateway:8080/demo-svc/hello
//
// 3) 查询服务列表：
//
//	curl http://gateway:8080/
//
// # 服务路由
//
// Gateway 通过 URL 路径中的服务名称路由请求:
//
//	HTTP 请求: http://gateway:8080/{service_name}/api/v1/users
//	Debug 请求: http://gateway:6060/{service_name}/debug/pprof
//
// 如果不指定服务名称，Gateway 返回已注册服务列表:
//
//	curl http://gateway:8080/
//	{
//	  "services": [
//	    {"name": "user-service", "version": "v1.0.0", ...},
//	    {"name": "order-service", "version": "v2.0.0", ...}
//	  ],
//	  "count": 2
//	}
//
// # 查询服务
//
// 获取已注册的服务:
//
//	// 获取所有服务
//	services := gw.Services()
//	for _, svc := range services {
//		fmt.Printf("Service: %s, Version: %s\n", svc.Name, svc.Version)
//	}
//
//	// 获取特定服务
//	svc, err := gw.GetService("my-service")
//	if err != nil {
//		log.Printf("Service not found: %v", err)
//	}
//
// # 传输协议
//
// 支持多种传输协议，需要与 Agent 使用相同的协议:
//
//   - yamux: 基于 TCP 的多路复用（默认）
//   - quic: 基于 UDP 的 QUIC 协议
//   - kcp: 基于 UDP 的 KCP 协议
//
// # TLS 配置
//
// 通过 TransportOptions 配置 TLS:
//
//	gw := tunnelgateway.New(&tunnelgateway.Config{
//		ListenAddr: ":9000",
//		Transport:  "yamux",
//		TransportOptions: &tunnelgateway.TransportOptions{
//			EnableTLS: true,
//			CertFile:  "/path/to/server.crt",
//			KeyFile:   "/path/to/server.key",
//			CAFile:    "/path/to/ca.crt",  // 用于验证客户端证书
//		},
//		HTTPPort: 8080,
//	})
//
// # 状态监控
//
// 检查 Gateway 状态:
//
//	status := gw.Status()
//	switch status {
//	case tunnelgateway.StatusRunning:
//		fmt.Println("Gateway 运行中")
//	case tunnelgateway.StatusStopped:
//		fmt.Println("Gateway 已停止")
//	}
//
// # WebSocket 支持
//
// Gateway 自动支持 WebSocket 请求的代理:
//
//	// Agent 端运行 WebSocket 服务
//	// Gateway 会自动检测并处理 WebSocket 升级请求
//	ws://gateway:8080/{service_name}/ws
//
// # 健康检查
//
// Gateway 自动进行服务健康检查，移除不健康的服务。
// 健康检查间隔可通过配置调整:
//
//	gw := tunnelgateway.New(&tunnelgateway.Config{
//		ListenAddr:          ":9000",
//		Transport:           "yamux",
//		HealthCheckInterval: 30,  // 健康检查间隔（秒）
//	})
package tunnelgateway
