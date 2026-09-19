package gateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// invokeFnConn is a Backend that runs a custom unary Invoke.
type invokeFnConn struct {
	fn func(ctx context.Context, method string, in, out any) error
}

func (f *invokeFnConn) Invoke(ctx context.Context, method string, in, out any, _ ...grpc.CallOption) error {
	if f.fn == nil {
		return nil
	}
	return f.fn(ctx, method, in, out)
}

func (f *invokeFnConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, status.Error(codes.Unimplemented, "stream not supported in test")
}

func encodeGRPCWebFrame(msg proto.Message) []byte {
	raw, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	frame := make([]byte, 5+len(raw))
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(raw)))
	copy(frame[5:], raw)
	return frame
}

func parseGRPCWebResponse(t *testing.T, raw []byte) (messages [][]byte, trailer http.Header) {
	t.Helper()
	trailer = make(http.Header)
	off := 0
	for off+5 <= len(raw) {
		flags := raw[off]
		n := int(binary.BigEndian.Uint32(raw[off+1 : off+5]))
		off += 5
		if off+n > len(raw) {
			t.Fatalf("truncated frame at %d: need %d have %d", off-5, n, len(raw)-(off-5))
		}
		payload := raw[off : off+n]
		off += n
		if flags&0x80 != 0 {
			// HTTP/1.1 header block
			for _, line := range strings.Split(string(payload), "\r\n") {
				if line == "" {
					continue
				}
				k, v, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				trailer.Add(strings.TrimSpace(k), strings.TrimSpace(v))
			}
			continue
		}
		messages = append(messages, append([]byte(nil), payload...))
	}
	return messages, trailer
}

func setupGRPCWebEchoMux(t *testing.T, invoke func(ctx context.Context, method string, in, out any) error) (*Mux, *fiber.App) {
	t.Helper()
	mux := NewMux()
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.StringValue")
	if err != nil {
		t.Fatalf("input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.StringValue")
	if err != nil {
		t.Fatalf("output type: %v", err)
	}

	fullMethod := "/test.v1.Echo/Echo"
	method := &methodWrapper{
		srv: &serviceWrapper{
			opts:           mux.opts,
			remoteProxyCli: &invokeFnConn{fn: invoke},
		},
		grpcFullMethod: fullMethod,
		inputType:      inType,
		outputType:     outType,
	}
	mux.opts.handlers[fullMethod] = method
	if err = mux.routerTree.Add("POST", fullMethod, fullMethod, nil); err != nil {
		t.Fatalf("add route: %v", err)
	}

	app := fiber.New()
	app.All("/*", mux.Handler)
	return mux, app
}

func echoStringInvoke(_ context.Context, _ string, in, out any) error {
	req, ok := in.(*wrapperspb.StringValue)
	if !ok {
		return status.Errorf(codes.Internal, "bad in type %T", in)
	}
	rsp, ok := out.(*wrapperspb.StringValue)
	if !ok {
		return status.Errorf(codes.Internal, "bad out type %T", out)
	}
	rsp.Value = "echo:" + req.GetValue()
	return nil
}

func TestGRPCWeb_BinaryUnaryRoundTrip(t *testing.T) {
	_, app := setupGRPCWebEchoMux(t, echoStringInvoke)

	reqBody := encodeGRPCWebFrame(wrapperspb.String("hello"))
	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.Echo/Echo", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("Accept", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d body=%q", resp.StatusCode, body)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/grpc-web+proto" {
		t.Fatalf("content-type=%q", ct)
	}

	msgs, trailer := parseGRPCWebResponse(t, body)
	if len(msgs) != 1 {
		t.Fatalf("messages=%d trailer=%v raw=%v", len(msgs), trailer, body)
	}
	var out wrapperspb.StringValue
	if err := proto.Unmarshal(msgs[0], &out); err != nil {
		t.Fatal(err)
	}
	if out.GetValue() != "echo:hello" {
		t.Fatalf("value=%q", out.GetValue())
	}
	if trailer.Get("Grpc-Status") != "0" && trailer.Get("grpc-status") != "0" {
		t.Fatalf("trailer=%v", trailer)
	}
}

func TestGRPCWeb_TextUnaryRoundTrip(t *testing.T) {
	_, app := setupGRPCWebEchoMux(t, echoStringInvoke)

	frame := encodeGRPCWebFrame(wrapperspb.String("text-ok"))
	reqBody := base64.StdEncoding.EncodeToString(frame)
	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.Echo/Echo", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/grpc-web-text+proto")
	req.Header.Set("Accept", "application/grpc-web-text+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	b64, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d body=%q", resp.StatusCode, b64)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/grpc-web-text+proto" {
		t.Fatalf("content-type=%q", ct)
	}

	// Response must be one continuous base64 stream (client atob / DecodeString).
	raw, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		t.Fatalf("response not valid base64 (%q): %v", b64, err)
	}

	msgs, trailer := parseGRPCWebResponse(t, raw)
	if len(msgs) != 1 {
		t.Fatalf("messages=%d trailer=%v", len(msgs), trailer)
	}
	var out wrapperspb.StringValue
	if err := proto.Unmarshal(msgs[0], &out); err != nil {
		t.Fatal(err)
	}
	if out.GetValue() != "echo:text-ok" {
		t.Fatalf("value=%q", out.GetValue())
	}
	if got := trailer.Get("grpc-status"); got != "0" && trailer.Get("Grpc-Status") != "0" {
		t.Fatalf("trailer=%v", trailer)
	}
}

func TestGRPCWeb_TextUnaryErrorTrailer(t *testing.T) {
	_, app := setupGRPCWebEchoMux(t, func(context.Context, string, any, any) error {
		return status.Error(codes.InvalidArgument, "bad name")
	})

	frame := encodeGRPCWebFrame(wrapperspb.String("x"))
	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.Echo/Echo",
		strings.NewReader(base64.StdEncoding.EncodeToString(frame)))
	req.Header.Set("Content-Type", "application/grpc-web-text+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	b64, _ := io.ReadAll(resp.Body)
	raw, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		t.Fatalf("error response must still be valid base64 (%q): %v", b64, err)
	}
	msgs, trailer := parseGRPCWebResponse(t, raw)
	if len(msgs) != 0 {
		t.Fatalf("expected no data frames on error, got %d", len(msgs))
	}
	code := trailer.Get("grpc-status")
	if code == "" {
		code = trailer.Get("Grpc-Status")
	}
	if code != "3" { // InvalidArgument
		t.Fatalf("grpc-status=%q trailer=%v", code, trailer)
	}
	if vals := trailer.Values("grpc-status"); len(vals) > 1 {
		t.Fatalf("duplicate grpc-status values: %v", vals)
	}
	if vals := trailer.Values("Grpc-Status"); len(vals) > 1 {
		t.Fatalf("duplicate Grpc-Status values: %v", vals)
	}
	msg := trailer.Get("grpc-message")
	if msg == "" {
		msg = trailer.Get("Grpc-Message")
	}
	if !strings.Contains(msg, "bad") && !strings.Contains(msg, "name") {
		t.Fatalf("grpc-message=%q", msg)
	}
}

func TestGRPCWeb_TextInvalidBase64Request(t *testing.T) {
	_, app := setupGRPCWebEchoMux(t, echoStringInvoke)

	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.Echo/Echo", strings.NewReader("!!!not-base64!!!"))
	req.Header.Set("Content-Type", "application/grpc-web-text+proto")
	req.Header.Set("X-Grpc-Web", "1")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == fiber.StatusOK {
		t.Fatalf("expected non-OK for invalid base64, status=%d body=%q", resp.StatusCode, body)
	}
	if len(body) == 0 {
		t.Fatal("expected error payload")
	}
}

func TestDecodeGRPCWebTextBody_RoundTrip(t *testing.T) {
	app := fiber.New()
	fctx := fasthttpRequestCtxWithBody(t, base64.StdEncoding.EncodeToString([]byte{0, 0, 0, 0, 3, 'a', 'b', 'c'}))
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	if err := decodeGRPCWebTextBody(ctx); err != nil {
		t.Fatal(err)
	}
	got := ctx.Body()
	want := []byte{0, 0, 0, 0, 3, 'a', 'b', 'c'}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDecodeGRPCWebTextBody_Invalid(t *testing.T) {
	app := fiber.New()
	fctx := fasthttpRequestCtxWithBody(t, "@@@@")
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	if err := decodeGRPCWebTextBody(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestPrepareGRPCWeb_RewritesContentType(t *testing.T) {
	app := fiber.New()
	fctx := fasthttpRequestCtxWithBody(t, base64.StdEncoding.EncodeToString(encodeGRPCWebFrame(wrapperspb.String("x"))))
	fctx.Request.Header.SetMethod("POST")
	fctx.Request.Header.SetContentType("application/grpc-web-text+proto")
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	f := newHTTPFrontend(NewMux())
	typ, enc, handled, err := f.prepareGRPCWeb(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected grpc-web-text handled")
	}
	if ct := string(ctx.Request().Header.ContentType()); ct != "application/grpc+proto" {
		t.Fatalf("rewritten content-type=%q", ct)
	}
	if typ != grpcWebText || enc != "proto" {
		t.Fatalf("typ=%q enc=%q", typ, enc)
	}
}

func TestFiberWebWriter_TextDataAndTrailerDecodableByAtob(t *testing.T) {
	// Regression: per-Write padded base64 broke browser atob on the full body.
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	var buf bytes.Buffer
	bw := newBase64ChunkWriter(&buf)
	w := &fiberWebWriter{ctx: ctx, typ: grpcWebText, enc: "proto", resp: bw, respCloser: bw}

	payload := wrapperspb.String("payload-bytes-xx") // length not multiple of 3 after framing
	frame := encodeGRPCWebFrame(payload)
	if _, err := w.Write(frame); err != nil {
		t.Fatal(err)
	}
	applyGRPCWebMetadata(ctx, metadata.Pairs("grpc-status", "0", "grpc-message", "ok"))
	w.flushWithTrailer()

	encoded := buf.String()
	if strings.Contains(encoded, "==") && strings.Count(encoded, "=") > 2 {
		// Streaming encoder should pad only once at the end, not after every Write.
		// (Multiple '=' only at end is fine; mid-stream padding would be '==' then more alphabet.)
		idx := strings.Index(encoded, "=")
		if idx >= 0 && idx < len(encoded)-2 {
			rest := encoded[idx:]
			if strings.ContainsAny(rest, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/") {
				t.Fatalf("padding appeared mid-stream: %q", encoded)
			}
		}
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("atob-equivalent decode failed: %v body=%q", err, encoded)
	}
	msgs, trailer := parseGRPCWebResponse(t, raw)
	if len(msgs) != 1 {
		t.Fatalf("msgs=%d", len(msgs))
	}
	var out wrapperspb.StringValue
	if err := proto.Unmarshal(msgs[0], &out); err != nil {
		t.Fatal(err)
	}
	if out.GetValue() != payload.GetValue() {
		t.Fatalf("value=%q", out.GetValue())
	}
	if trailer.Get("grpc-status") != "0" && trailer.Get("Grpc-Status") != "0" {
		t.Fatalf("trailer=%v", trailer)
	}
}

// fasthttpRequestCtxWithBody builds a RequestCtx with a buffered body (no BodyStream).
func fasthttpRequestCtxWithBody(t *testing.T, body string) *fasthttp.RequestCtx {
	t.Helper()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("POST")
	fctx.Request.SetBodyString(body)
	return fctx
}
