// Command verify exercises every gateway frontend exposed by the grpcwebsocket
// example. Run the example server first:
//
//	go run ./internal/examples/grpcwebsocket/
//	go run ./internal/examples/grpcwebsocket/verify/
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	greeterpb "github.com/pubgo/lava/v2/internal/examples/grpcweb/proto"
)

func main() {
	verifyWebSocketUnary()
	verifyWebSocketRESTPath()
	verifyWebSocketServerStream()
	verifyWebSocketBidi()
	verifyNativeGRPC()
	log.Println("ALL CHECKS PASSED")
}

func verifyWebSocketUnary() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx,
		"ws://localhost:8081/grpcweb.example.v1.GreeterService/SayHello",
		&websocket.DialOptions{Subprotocols: []string{"grpc-ws-json"}},
	)
	if err != nil {
		log.Fatalf("ws unary dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	if err = conn.Write(ctx, websocket.MessageText, []byte(`{"name":"WS"}`)); err != nil {
		log.Fatalf("ws unary write: %v", err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		log.Fatalf("ws unary read: %v", err)
	}
	log.Printf("[websocket unary] %s", string(data))

	_, _, err = conn.Read(ctx)
	var ce websocket.CloseError
	if !errors.As(err, &ce) {
		log.Fatalf("ws expected close error, got: %v", err)
	}
	var payload struct {
		GRPCStatus  uint32 `json:"grpcStatus"`
		GRPCMessage string `json:"grpcMessage"`
	}
	if jErr := json.Unmarshal([]byte(ce.Reason), &payload); jErr != nil {
		log.Fatalf("ws close reason not structured JSON: %q (%v)", ce.Reason, jErr)
	}
	if payload.GRPCStatus != 0 {
		log.Fatalf("ws expected grpcStatus=0, got %d", payload.GRPCStatus)
	}
	log.Printf("[websocket close] code=%d grpcStatus=%d grpcMessage=%q",
		ce.Code, payload.GRPCStatus, payload.GRPCMessage)
}

func verifyWebSocketRESTPath() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx,
		"ws://localhost:8081/v1/greeter/hello",
		&websocket.DialOptions{Subprotocols: []string{"grpc-ws-json"}},
	)
	if err != nil {
		log.Fatalf("ws REST path dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	if err = conn.Write(ctx, websocket.MessageText, []byte(`{"name":"REST-WS"}`)); err != nil {
		log.Fatalf("ws REST path write: %v", err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		log.Fatalf("ws REST path read: %v", err)
	}
	log.Printf("[websocket REST path] %s", string(data))
}

func verifyWebSocketServerStream() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx,
		"ws://localhost:8081/grpcweb.example.v1.GreeterService/WatchHello",
		&websocket.DialOptions{Subprotocols: []string{"grpc-ws-json"}},
	)
	if err != nil {
		log.Fatalf("ws server-stream dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	if err = conn.Write(ctx, websocket.MessageText, []byte(`{"name":"WS","count":3}`)); err != nil {
		log.Fatalf("ws server-stream write: %v", err)
	}

	for i := 1; i <= 3; i++ {
		_, data, rErr := conn.Read(ctx)
		if rErr != nil {
			log.Fatalf("ws server-stream read #%d: %v", i, rErr)
		}
		log.Printf("[websocket server-stream] frame #%d %s", i, string(data))
	}
}

func verifyWebSocketBidi() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx,
		"ws://localhost:8081/grpcweb.example.v1.GreeterService/Chat",
		&websocket.DialOptions{Subprotocols: []string{"grpc-ws-json"}},
	)
	if err != nil {
		log.Fatalf("ws bidi dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	for _, text := range []string{"hi", "there"} {
		if err = conn.Write(ctx, websocket.MessageText, []byte(`{"text":"`+text+`"}`)); err != nil {
			log.Fatalf("ws bidi write: %v", err)
		}
		_, data, rErr := conn.Read(ctx)
		if rErr != nil {
			log.Fatalf("ws bidi read: %v", rErr)
		}
		log.Printf("[websocket bidi] sent=%q recv=%s", text, string(data))
	}
	_ = conn.Close(websocket.StatusNormalClosure, "")
}

func verifyNativeGRPC() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("native gRPC dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cli := greeterpb.NewGreeterServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.SayHello(ctx, &greeterpb.HelloRequest{Name: "NATIVE"})
	if err != nil {
		log.Fatalf("native gRPC SayHello: %v", err)
	}
	log.Printf("[native gRPC unary] %s", resp.GetMessage())

	stream, err := cli.Chat(ctx)
	if err != nil {
		log.Fatalf("native gRPC Chat open: %v", err)
	}
	for _, text := range []string{"hi", "there"} {
		if err = stream.Send(&greeterpb.ChatMessage{Text: text}); err != nil {
			log.Fatalf("native gRPC Chat send: %v", err)
		}
		msg, recvErr := stream.Recv()
		if recvErr != nil {
			log.Fatalf("native gRPC Chat recv: %v", recvErr)
		}
		log.Printf("[native gRPC bidi] sent=%q recv=%q", text, msg.GetText())
	}
	_ = stream.CloseSend()
}
