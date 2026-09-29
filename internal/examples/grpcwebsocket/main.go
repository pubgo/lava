// Package main 提供 Gateway 多协议示例：同一套 gRPC handler 同时暴露
// HTTP/REST + gRPC-Web、WebSocket、Native gRPC。
//
// 运行:
//
//	go run ./internal/examples/grpcwebsocket
//
// 验证（需先启动本服务）:
//
//	go run ./internal/examples/grpcwebsocket/verify/
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
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	greeterpb "github.com/pubgo/lava/v2/internal/examples/grpcweb/proto"
	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/servers/gatewayserver"
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
		Message:   "Hello, " + name + "!",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *greeterService) SayGoodbye(_ context.Context, req *greeterpb.GoodbyeRequest) (*greeterpb.GoodbyeResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	return &greeterpb.GoodbyeResponse{
		Message:   "Goodbye, " + name + "!",
		Timestamp: time.Now().Unix(),
	}, nil
}

func firstMD(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (s *greeterService) WatchHello(req *greeterpb.WatchHelloRequest, stream greeterpb.GreeterService_WatchHelloServer) error {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	// count==0 → subscribe until client cancels; count>0 → finite pushes.
	count := req.GetCount()
	in, _ := metadata.FromIncomingContext(stream.Context())
	// Explicit header wins (frontend sends this in subscribe mode).
	if firstMD(in, "x-demo-subscribe") == "1" {
		count = 0
	}
	interval := time.Second
	if v := firstMD(in, "x-demo-interval-ms"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 200 {
			interval = time.Duration(ms) * time.Millisecond
		}
	}
	mode := "finite"
	if count <= 0 {
		mode = "subscribe"
	}
	log.Printf("WatchHello name=%s mode=%s count=%d interval=%s", name, mode, count, interval)

	if err := stream.SendHeader(metadata.Pairs(
		"x-demo-echo", firstMD(in, "x-demo-token"),
		"x-demo-method", "WatchHello",
		"x-demo-stream-mode", mode,
		"x-demo-interval-ms", strconv.FormatInt(interval.Milliseconds(), 10),
		"x-demo-stream-count", fmt.Sprintf("%d", count),
	)); err != nil {
		return err
	}
	stream.SetTrailer(metadata.Pairs("x-demo-trailer", "stream-done"))

	ctx := stream.Context()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := int32(1); ; i++ {
		if count > 0 && i > count {
			return nil
		}
		if err := stream.Send(&greeterpb.HelloResponse{
			Message:   fmt.Sprintf("Hello, %s! push #%d", name, i),
			Timestamp: time.Now().Unix(),
		}); err != nil {
			return err
		}
		if count > 0 && i == count {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
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

	// RPC middleware wraps the full Dispatch for every frontend (local + proxy).
	mux.UseRPCMiddleware(func(ctx context.Context, op *gateway.Operation, next gateway.RPCHandler) (metadata.MD, metadata.MD, error) {
		h, t, err := next(ctx)
		if err != nil {
			return h, t, err
		}
		if h == nil {
			h = metadata.MD{}
		}
		h.Set("x-example-mw", "1")
		if op != nil {
			h.Set("x-example-op", op.FullMethod)
		}
		return h, t, nil
	})

	surface := gatewayserver.NewGatewaySurface(mux, gateway.WSOptionsFromConfig(gateway.WSConfig{
		InsecureSkipVerify: true,
	})...)

	grpcLis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(surface.GRPCServerOptions...)
	go func() {
		log.Println("Native gRPC server listening on :50051")
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatal(err)
		}
	}()

	wsServer := &http.Server{
		Addr:              ":8081",
		Handler:           surface.WebSocketHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Println("WebSocket server listening on :8081")
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	app := fiber.New(fiber.Config{AppName: "Gateway Multi-Protocol Example"})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Content-Type", "X-Grpc-Web", "X-User-Agent",
			"Grpc-Encoding", "Grpc-Accept-Encoding",
			"X-Demo-Token", "X-Request-Id", "X-Demo-Interval-Ms", "X-Demo-Subscribe",
		},
		ExposeHeaders: []string{
			"Grpc-Status", "Grpc-Message",
			"Grpc-Encoding", "Grpc-Accept-Encoding",
			"X-Example-Mw", "X-Example-Op",
			"X-Demo-Echo", "X-Demo-Method", "X-Demo-Trailer",
			"X-Demo-Stream-Count", "X-Demo-Stream-Mode", "X-Demo-Interval-Ms",
		},
	}))

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	app.Use("/", static.New("", static.Config{FS: staticFS, Browse: true}))
	app.All("/v1/*", surface.FiberHandler)
	app.Post("/grpcweb.example.v1.GreeterService/*", surface.FiberHandler)

	log.Println("HTTP/gRPC-Web listening on :8080")
	log.Println("  REST:     POST /v1/greeter/hello")
	log.Println("  gRPC-Web: POST /grpcweb.example.v1.GreeterService/SayHello")
	log.Println("  WS:       ws://localhost:8081/grpcweb.example.v1.GreeterService/SayHello")
	log.Println("  Native:   localhost:50051")
	log.Println("Verify: go run ./internal/examples/grpcwebsocket/verify/")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
