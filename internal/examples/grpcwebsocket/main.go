// Package main 提供 Gateway WebSocket 示例服务。
//
// 本示例演示如何通过 coder/websocket 前端调用已注册的 gRPC handler。
// WebSocket 前端运行在标准 net/http 栈上，与 Fiber 上的 HTTP/gRPC-Web 前端并存。
//
// 运行:
//
//	go run ./internal/examples/grpcwebsocket
//
// 测试:
//
//  1. 浏览器打开 http://localhost:8080/
//  2. 点击 SayHello / SayGoodbye 按钮，WebSocket 连接 ws://localhost:8081/...
//  3. curl 无法直接测试 WS，请使用浏览器或 wscat
//
//  4. 原生 gRPC 客户端连接 localhost:50051（与 HTTP/WS 共享同一套 handler）
package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"

	greeterpb "github.com/pubgo/lava/v2/internal/examples/grpcweb/proto"
	"github.com/pubgo/lava/v2/pkg/gateway"
)

//go:embed static/*
var staticFiles embed.FS

type greeterService struct {
	greeterpb.UnimplementedGreeterServiceServer
}

func (s *greeterService) SayHello(_ context.Context, req *greeterpb.HelloRequest) (*greeterpb.HelloResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	return &greeterpb.HelloResponse{
		Message:   "Hello, " + name + "! (via WebSocket)",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *greeterService) SayGoodbye(_ context.Context, req *greeterpb.GoodbyeRequest) (*greeterpb.GoodbyeResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	return &greeterpb.GoodbyeResponse{
		Message:   "Goodbye, " + name + "! (via WebSocket)",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *greeterService) WatchHello(req *greeterpb.WatchHelloRequest, stream greeterpb.GreeterService_WatchHelloServer) error {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	count := req.GetCount()
	if count <= 0 {
		count = 3
	}
	for i := int32(1); i <= count; i++ {
		if err := stream.Send(&greeterpb.HelloResponse{
			Message:   fmt.Sprintf("Hello, %s! stream #%d (via WebSocket)", name, i),
			Timestamp: time.Now().Unix(),
		}); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (s *greeterService) Chat(stream greeterpb.GreeterService_ChatServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		text := strings.TrimSpace(msg.GetText())
		if strings.EqualFold(text, "bye") {
			return stream.Send(&greeterpb.ChatMessage{Text: "bye (server closing)"})
		}
		if err = stream.Send(&greeterpb.ChatMessage{Text: "echo: " + text}); err != nil {
			return err
		}
	}
}

func main() {
	mux := gateway.NewMux()
	mux.RegisterService(&greeterpb.GreeterService_ServiceDesc, &greeterService{})

	// Native gRPC 前端：透传到 Mux（RegisterService 一次，多协议复用）
	grpcLis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(mux.GRPCServerOptions()...)
	go func() {
		log.Println("Native gRPC server listening on :50051")
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatal(err)
		}
	}()

	// WebSocket 前端：独立 net/http 端口
	wsHandler := mux.WebSocketHandler(gateway.WSOptionsFromConfig(gateway.WSConfig{
		InsecureSkipVerify: true,
	})...)
	wsServer := &http.Server{
		Addr:              ":8081",
		Handler:           wsHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Println("WebSocket server listening on :8081")
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// HTTP/REST + gRPC-Web：Fiber 端口
	app := fiber.New(fiber.Config{AppName: "Gateway WebSocket Example"})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
	}))

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	app.Use("/", static.New("", static.Config{FS: staticFS, Browse: true}))
	app.All("/v1/*", mux.Handler)
	app.Post("/grpcweb.example.v1.GreeterService/*", mux.Handler)

	log.Println("HTTP server listening on :8080")
	log.Println("Open http://localhost:8080/ to test WebSocket gateway")
	log.Println("WebSocket endpoint example:")
	log.Println("  ws://localhost:8081/grpcweb.example.v1.GreeterService/SayHello")
	log.Println("  ws://localhost:8081/grpcweb.example.v1.GreeterService/WatchHello  (server-stream)")
	log.Println("  ws://localhost:8081/grpcweb.example.v1.GreeterService/Chat       (bidi)")
	log.Println("Native gRPC endpoint: localhost:50051")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
