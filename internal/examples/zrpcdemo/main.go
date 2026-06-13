package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/signals"
	zrpcdemov1 "github.com/pubgo/lava/v2/pkg/proto/zrpcdemov1"
	"github.com/pubgo/lava/v2/pkg/zrpc"
	"github.com/pubgo/lava/v2/servers/zrpcs"
)

// 运行方式：
//   1. 先启动 nats-server（默认 nats://127.0.0.1:4222）
//   2. go run ./internal/examples/zrpcdemo
//   3. 使用 zrpcc、生成客户端或 zrpccli 调用：
//      - svc.zrpcdemo.EchoService/Echo
//      - svc.zrpcdemo.v1.EchoService/Reverse

type echoService struct{}

func (echoService) Echo(_ context.Context, req *zrpcdemov1.EchoRequest) (*zrpcdemov1.EchoResponse, error) {
	return &zrpcdemov1.EchoResponse{Message: req.GetMessage()}, nil
}

func (echoService) Reverse(_ context.Context, req *zrpcdemov1.EchoRequest) (*zrpcdemov1.EchoResponse, error) {
	runes := []rune(req.GetMessage())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return &zrpcdemov1.EchoResponse{Message: string(runes)}, nil
}

func (echoService) EchoStream(_ context.Context, req *zrpcdemov1.EchoRequest, stream zrpcdemov1.EchoService_EchoStreamZrpcServerStream) error {
	base := req.GetMessage()
	for _, val := range []string{base, strings.ToUpper(base), strings.ToLower(base)} {
		if err := stream.Send(&zrpcdemov1.EchoResponse{Message: val}); err != nil {
			return err
		}
	}

	return nil
}

func (echoService) Collect(_ context.Context, stream zrpcdemov1.EchoService_CollectZrpcServerStream) (*zrpcdemov1.EchoResponse, error) {
	parts := make([]string, 0, 4)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		parts = append(parts, req.GetMessage())
	}

	return &zrpcdemov1.EchoResponse{Message: strings.Join(parts, ",")}, nil
}

func (echoService) Chat(_ context.Context, stream zrpcdemov1.EchoService_ChatZrpcServerStream) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if err = stream.Send(&zrpcdemov1.EchoResponse{Message: "chat:" + req.GetMessage()}); err != nil {
			return err
		}
	}
}

func main() {
	defer recovery.Exit()

	logger := log.GetLogger().WithName("zrpcdemo")
	metric := tally.NoopScope

	url := os.Getenv("NATS_URL")
	if strings.TrimSpace(url) == "" {
		url = "nats://127.0.0.1:4222"
	}

	svc := zrpcs.New(zrpcs.Params{
		Log:    logger,
		Metric: metric,
		Conf:   &zrpcs.Config{URL: url},
		Registers: []zrpcs.RegisterFunc{
			func(srv *zrpc.Server) error {
				return zrpcdemov1.RegisterEchoServiceZrpcRoutes(srv, echoService{}, "")
			},
		},
	})

	println("zrpc demo listening on", url)
	println("echo subject:", zrpcdemov1.EchoService_EchoSubject)
	println("reverse subject:", zrpcdemov1.EchoService_ReverseSubject)
	println("try payload: {\"message\":\"hello\"}")

	if err := svc.Serve(signals.Context()); err != nil {
		panic(err)
	}

	fmt.Println("zrpc demo stopped", strings.TrimSpace(url))
}
