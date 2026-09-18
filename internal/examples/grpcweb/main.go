// Package main 提供 gRPC Web 示例服务（HTTP/REST + gRPC-Web）。
//
// 多协议（含 WebSocket / Native gRPC）见 internal/examples/grpcwebsocket。
//
// 运行:
//
//	go run ./internal/examples/grpcweb
//
// 测试:
//
//  1. curl HTTP/JSON:
//     curl -X POST http://localhost:8080/v1/greeter/hello -H "Content-Type: application/json" -d '{"name":"World"}'
//
//  2. 浏览器打开 http://localhost:8080/
package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
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
)

//go:embed static/*
var staticFiles embed.FS

type greeterService struct {
	greeterpb.UnimplementedGreeterServiceServer
}

func firstMD(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (s *greeterService) SayHello(ctx context.Context, req *greeterpb.HelloRequest) (*greeterpb.HelloResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	in, _ := metadata.FromIncomingContext(ctx)
	_ = grpc.SendHeader(ctx, metadata.Pairs(
		"x-demo-echo", firstMD(in, "x-demo-token"),
		"x-demo-method", "SayHello",
	))
	_ = grpc.SetTrailer(ctx, metadata.Pairs("x-demo-trailer", "unary-done"))
	return &greeterpb.HelloResponse{
		Message:   "Hello, " + name + "!",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *greeterService) SayGoodbye(ctx context.Context, req *greeterpb.GoodbyeRequest) (*greeterpb.GoodbyeResponse, error) {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	in, _ := metadata.FromIncomingContext(ctx)
	_ = grpc.SendHeader(ctx, metadata.Pairs(
		"x-demo-echo", firstMD(in, "x-demo-token"),
		"x-demo-method", "SayGoodbye",
	))
	_ = grpc.SetTrailer(ctx, metadata.Pairs("x-demo-trailer", "unary-done"))
	return &greeterpb.GoodbyeResponse{
		Message:   "Goodbye, " + name + "! See you next time.",
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *greeterService) WatchHello(req *greeterpb.WatchHelloRequest, stream greeterpb.GreeterService_WatchHelloServer) error {
	name := req.GetName()
	if name == "" {
		name = "Anonymous"
	}
	// count==0 → keep pushing until the client cancels (pub/sub style).
	count := req.GetCount()
	in, _ := metadata.FromIncomingContext(stream.Context())
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
			// Client unsubscribed / aborted — normal for long-lived push.
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
	mux.UseRPCMiddleware(func(ctx context.Context, op *gateway.Operation, next gateway.RPCHandler) (metadata.MD, metadata.MD, error) {
		h, t, err := next(ctx)
		if err != nil {
			return h, t, err
		}
		if h == nil {
			h = metadata.MD{}
		}
		h.Set("x-example-mw", "1")
		return h, t, nil
	})

	app := fiber.New(fiber.Config{AppName: "gRPC Web Example"})
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
			"X-Example-Mw", "X-Demo-Echo", "X-Demo-Method",
			"X-Demo-Trailer", "X-Demo-Stream-Count",
			"X-Demo-Stream-Mode", "X-Demo-Interval-Ms",
		},
	}))

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	app.Use("/", static.New("", static.Config{FS: staticFS, Browse: true}))
	app.All("/v1/*", mux.Handler)
	app.Post("/grpcweb.example.v1.GreeterService/*", mux.Handler)

	log.Println("Starting gRPC Web Example Server on :8080")
	log.Println("Open http://localhost:8080/ in your browser")
	log.Println("For multi-protocol (WS + native gRPC) see ./internal/examples/grpcwebsocket")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
