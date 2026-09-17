// Command verify exercises every gateway frontend exposed by the grpcwebsocket
// example. Run the example server first:
//
//	go run ./internal/examples/grpcwebsocket/
//	go run ./internal/examples/grpcwebsocket/verify/
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	greeterpb "github.com/pubgo/lava/v2/internal/examples/grpcweb/proto"
)

func main() {
	verifyHTTPJSON()
	verifyHTTPClientStreamRejected()
	verifyGRPCWebUnary()
	verifyGRPCWebCompressed()
	verifyGRPCWebServerStream()
	verifyWebSocketUnary()
	verifyWebSocketRESTPath()
	verifyWebSocketServerStream()
	verifyWebSocketBidi()
	verifyNativeGRPC()
	log.Println("ALL CHECKS PASSED")
}

func verifyHTTPJSON() {
	resp, err := http.Post(
		"http://localhost:8080/v1/greeter/hello",
		"application/json",
		strings.NewReader(`{"name":"HTTP"}`),
	)
	if err != nil {
		log.Fatalf("http json: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("http json status=%d body=%s", resp.StatusCode, body)
	}
	if resp.Header.Get("X-Example-Mw") != "1" {
		log.Fatalf("http json missing x-example-mw header: %v", resp.Header)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Fatalf("http json decode: %v body=%s", err, body)
	}
	log.Printf("[http json] %s mw=%s op=%s", body, resp.Header.Get("X-Example-Mw"), resp.Header.Get("X-Example-Op"))
}

func verifyHTTPClientStreamRejected() {
	resp, err := http.Post(
		"http://localhost:8080/v1/greeter/chat",
		"application/json",
		strings.NewReader(`{"text":"hi"}`),
	)
	if err != nil {
		log.Fatalf("http chat: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		log.Fatalf("http chat should be rejected, body=%s", body)
	}
	var payload struct {
		Code    uint32 `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Fatalf("http chat decode: %v body=%s", err, body)
	}
	if codes.Code(payload.Code) != codes.Unimplemented {
		log.Fatalf("http chat want Unimplemented, got code=%d msg=%q", payload.Code, payload.Message)
	}
	log.Printf("[http chat rejected] code=%d msg=%q", payload.Code, payload.Message)
}

func verifyGRPCWebUnary() {
	reqMsg, err := proto.Marshal(&greeterpb.HelloRequest{Name: "WEB"})
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost,
		"http://localhost:8080/grpcweb.example.v1.GreeterService/SayHello",
		bytes.NewReader(grpcFrame(0, reqMsg)),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("grpc-web unary: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	msgs, trailer, _ := parseGRPCWeb(raw)
	if len(msgs) != 1 {
		log.Fatalf("grpc-web unary want 1 data frame, got %d trailer=%v raw=%x", len(msgs), trailer, raw)
	}
	if trailer.Get("Grpc-Status") != "0" && trailer.Get("grpc-status") != "0" {
		log.Fatalf("grpc-web unary trailer status=%q raw=%q", trailer.Get("grpc-status"), raw)
	}
	out := &greeterpb.HelloResponse{}
	if err := proto.Unmarshal(msgs[0], out); err != nil {
		log.Fatalf("grpc-web unary unmarshal: %v", err)
	}
	log.Printf("[grpc-web unary] %s trailer=%v", out.GetMessage(), trailer)
}

func verifyGRPCWebCompressed() {
	reqMsg, err := proto.Marshal(&greeterpb.HelloRequest{Name: "GZIP"})
	if err != nil {
		log.Fatal(err)
	}
	comp, err := gzipBytes(reqMsg)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost,
		"http://localhost:8080/grpcweb.example.v1.GreeterService/SayHello",
		bytes.NewReader(grpcFrame(1, comp)),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")
	req.Header.Set("Grpc-Encoding", "gzip")
	req.Header.Set("Grpc-Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("grpc-web gzip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	if enc := resp.Header.Get("Grpc-Encoding"); enc != "gzip" {
		log.Fatalf("grpc-web gzip response encoding=%q want gzip headers=%v", enc, resp.Header)
	}
	if accept := resp.Header.Get("Grpc-Accept-Encoding"); !strings.Contains(accept, "gzip") {
		log.Fatalf("grpc-web missing accept-encoding advertise: %q", accept)
	}

	msgs, trailer, compressed := parseGRPCWeb(raw)
	if len(msgs) != 1 {
		log.Fatalf("grpc-web gzip want 1 data frame, got %d", len(msgs))
	}
	if !compressed {
		log.Fatal("grpc-web gzip response data frame should be compressed")
	}
	if trailer.Get("Grpc-Status") != "0" && trailer.Get("grpc-status") != "0" {
		log.Fatalf("grpc-web gzip trailer status=%q", trailer.Get("grpc-status"))
	}
	plain, err := gunzipBytes(msgs[0])
	if err != nil {
		log.Fatalf("grpc-web gzip decompress: %v", err)
	}
	out := &greeterpb.HelloResponse{}
	if err := proto.Unmarshal(plain, out); err != nil {
		log.Fatalf("grpc-web gzip unmarshal: %v", err)
	}
	log.Printf("[grpc-web gzip] %s encoding=%s", out.GetMessage(), resp.Header.Get("Grpc-Encoding"))
}

func verifyGRPCWebServerStream() {
	reqMsg, err := proto.Marshal(&greeterpb.WatchHelloRequest{Name: "STREAM", Count: 2})
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost,
		"http://localhost:8080/grpcweb.example.v1.GreeterService/WatchHello",
		bytes.NewReader(grpcFrame(0, reqMsg)),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("grpc-web stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	msgs, trailer, _ := parseGRPCWeb(raw)
	if len(msgs) != 2 {
		log.Fatalf("grpc-web stream want 2 frames, got %d trailer=%v", len(msgs), trailer)
	}
	if trailer.Get("Grpc-Status") != "0" && trailer.Get("grpc-status") != "0" {
		log.Fatalf("grpc-web stream trailer status=%q", trailer.Get("grpc-status"))
	}
	log.Printf("[grpc-web server-stream] frames=%d status=%s", len(msgs), trailerStatus(trailer))
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

func grpcFrame(flags byte, payload []byte) []byte {
	frame := make([]byte, 5+len(payload))
	frame[0] = flags
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(payload)))
	copy(frame[5:], payload)
	return frame
}

func parseGRPCWeb(raw []byte) (msgs [][]byte, trailer http.Header, dataCompressed bool) {
	trailer = make(http.Header)
	for len(raw) >= 5 {
		flags := raw[0]
		n := binary.BigEndian.Uint32(raw[1:5])
		if uint32(len(raw)) < 5+n {
			break
		}
		payload := raw[5 : 5+n]
		raw = raw[5+n:]
		if flags&0x80 != 0 {
			_ = parseMIMEHeaders(payload, trailer)
			return msgs, trailer, dataCompressed
		}
		if flags&0x01 != 0 {
			dataCompressed = true
		}
		msgs = append(msgs, append([]byte(nil), payload...))
	}
	return msgs, trailer, dataCompressed
}

func parseMIMEHeaders(payload []byte, dst http.Header) error {
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(payload)))
	mime, err := tp.ReadMIMEHeader()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	for k, vals := range mime {
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
	return nil
}

func trailerStatus(h http.Header) string {
	if v := h.Get("Grpc-Status"); v != "" {
		return v
	}
	return h.Get("grpc-status")
}

func gzipBytes(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(in); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gunzipBytes(in []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}
