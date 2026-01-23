// Package main 提供 gRPC Web 示例服务
//
// 本示例用于验证和测试 gateway 的 gRPC Web 实现
//
// 运行方式:
//
//	go run ./internal/examples/grpcweb
//
// 测试方式:
//
//  1. 使用 curl 测试普通 HTTP/JSON:
//     curl -X POST http://localhost:8080/v1/greeter/hello -H "Content-Type: application/json" -d '{"name":"World"}'
//
//  2. 使用浏览器打开 http://localhost:8080/ 测试 gRPC Web
//
//  3. 使用 curl 测试 gRPC Web:
//     curl -X POST http://localhost:8080/grpcweb.example.v1.GreeterService/SayHello \
//     -H "Content-Type: application/grpc-web+proto" \
//     -d $'\x00\x00\x00\x00\x07\x0a\x05World' --output -
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"

	greeterpb "github.com/pubgo/lava/v2/internal/examples/grpcweb/proto"
	"github.com/pubgo/lava/v2/pkg/gateway"
)

//go:embed static/*
var staticFiles embed.FS

// greeterService 实现 GreeterServiceServer 接口
type greeterService struct {
	greeterpb.UnimplementedGreeterServiceServer
}

// SayHello 实现问候方法
func (s *greeterService) SayHello(ctx context.Context, req *greeterpb.HelloRequest) (*greeterpb.HelloResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	return &greeterpb.HelloResponse{
		Message:   "Hello, " + name + "!",
		Timestamp: time.Now().Unix(),
	}, nil
}

// SayGoodbye 实现告别方法
func (s *greeterService) SayGoodbye(ctx context.Context, req *greeterpb.GoodbyeRequest) (*greeterpb.GoodbyeResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	return &greeterpb.GoodbyeResponse{
		Message:   "Goodbye, " + name + "! See you next time.",
		Timestamp: time.Now().Unix(),
	}, nil
}

func main() {
	// 创建 Gateway Mux
	mux := gateway.NewMux()

	// 注册服务
	mux.RegisterService(&greeterpb.GreeterService_ServiceDesc, &greeterService{})

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		AppName: "gRPC Web Example",
	})

	// 添加中间件
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,POST,OPTIONS",
		AllowHeaders:  "Content-Type,X-Grpc-Web,X-User-Agent",
		ExposeHeaders: "Grpc-Status,Grpc-Message",
	}))

	// 静态文件服务
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(staticFS),
		Browse: true,
	}))

	// 注册 Gateway Handler
	app.All("/v1/*", mux.Handler)
	// 注册 gRPC Web 路由 (支持直接使用 gRPC 方法路径)
	app.Post("/grpcweb.example.v1.GreeterService/*", mux.Handler)

	log.Println("Starting gRPC Web Example Server on :8080")
	log.Println("Open http://localhost:8080/ in your browser to test gRPC Web")
	log.Println("Test with curl:")
	log.Println("  HTTP/JSON: curl -X POST http://localhost:8080/v1/greeter/hello -H 'Content-Type: application/json' -d '{\"name\":\"World\"}'")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
