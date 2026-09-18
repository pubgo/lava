package gateway

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHTTPFrontend_RejectsClientStreaming(t *testing.T) {
	assertHTTPRejectsClientStreams(t, false, "/v1/upload", "/test.v1.StreamService/Upload")
}

func TestHTTPFrontend_RejectsBidiStreaming(t *testing.T) {
	assertHTTPRejectsClientStreams(t, true, "/v1/chat", "/test.v1.StreamService/Chat")
}

func assertHTTPRejectsClientStreams(t *testing.T, serverStreams bool, path, fullMethod string) {
	t.Helper()

	mux := NewMux()
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	method := &methodWrapper{
		srv: &serviceWrapper{opts: mux.opts},
		grpcStreamDesc: &grpc.StreamDesc{
			ServerStreams: serverStreams,
			ClientStreams: true,
		},
		grpcFullMethod: fullMethod,
		inputType:      inType,
		outputType:     outType,
	}
	mux.opts.handlers[method.grpcFullMethod] = method
	if err = mux.routerTree.Add("POST", path, method.grpcFullMethod, nil); err != nil {
		t.Fatalf("add route: %v", err)
	}

	app := fiber.New()
	app.All("/*", mux.Handler)

	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusNotImplemented {
		t.Fatalf("want HTTP 501, got %d body=%q", resp.StatusCode, body)
	}
	var payload struct {
		Code    uint32 `json:"code"`
		Message string `json:"message"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json body: %v (%q)", err, body)
	}
	if codes.Code(payload.Code) != codes.Unimplemented {
		t.Fatalf("want Unimplemented code, got %d message=%q", payload.Code, payload.Message)
	}
	if !strings.Contains(payload.Message, "does not support client-streaming") {
		t.Fatalf("unexpected message: %q", payload.Message)
	}
}

func TestHTTPFrontend_MapsNotFound(t *testing.T) {
	mux := NewMux()
	app := fiber.New()
	app.All("/*", mux.Handler)

	req := httptest.NewRequest(fiber.MethodGet, "/no/such/route", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("want HTTP 404, got %d body=%q", resp.StatusCode, body)
	}
	var payload struct {
		Code uint32 `json:"code"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json body: %v (%q)", err, body)
	}
	if codes.Code(payload.Code) != codes.NotFound {
		t.Fatalf("want NotFound code, got %d", payload.Code)
	}
}

func TestWriteNDJSONStreamError_EmitsJSONErrorLine(t *testing.T) {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	writeNDJSONStreamError(bw, status.Error(codes.InvalidArgument, "bad id"))

	var payload struct {
		Code    uint32 `json:"code"`
		Message string `json:"message"`
	}
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("json body: %v (%q)", err, line)
	}
	if codes.Code(payload.Code) != codes.InvalidArgument || payload.Message != "bad id" {
		t.Fatalf("payload=%+v", payload)
	}
}

func TestHTTPFrontend_MapsInvalidArgumentFromHandler(t *testing.T) {
	mux := NewMux()
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	method := &methodWrapper{
		srv: &serviceWrapper{
			opts: mux.opts,
			remoteProxyCli: &fakeClientConn{
				invokeErr: status.Error(codes.InvalidArgument, "bad id"),
			},
		},
		grpcFullMethod: "/test.v1.Echo/Ping",
		inputType:      inType,
		outputType:     outType,
	}
	mux.opts.handlers[method.grpcFullMethod] = method
	if err = mux.routerTree.Add("POST", "/v1/ping", method.grpcFullMethod, nil); err != nil {
		t.Fatalf("add route: %v", err)
	}

	app := fiber.New()
	app.All("/*", mux.Handler)

	req := httptest.NewRequest(fiber.MethodPost, "/v1/ping", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("want HTTP 400, got %d body=%q", resp.StatusCode, body)
	}
}

func TestStreamHTTP_RecvMsg_SecondCallEOF(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Request().SetBody([]byte("{}"))

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}

	s := &streamHTTP{
		handler: ctx,
		method: &methodWrapper{
			srv:       &serviceWrapper{opts: NewMux().opts},
			inputType: inType,
		},
	}

	msg := &emptypb.Empty{}
	if err = s.RecvMsg(msg); err != nil {
		t.Fatalf("first RecvMsg: %v", err)
	}
	if err = s.RecvMsg(msg); err != io.EOF {
		t.Fatalf("second RecvMsg want EOF, got %v", err)
	}
}

func TestGetRouterTarget_MatchesMethodAndPath(t *testing.T) {
	mux := NewMux()
	if err := mux.routerTree.Add("GET", "/v1/users/{id}", "/test.v1.User/Get", nil); err != nil {
		t.Fatalf("add route: %v", err)
	}

	got, err := GetRouterTarget(mux, "GET", "/v1/users/42")
	if err != nil {
		t.Fatalf("GetRouterTarget: %v", err)
	}
	if got.Operation != "/test.v1.User/Get" {
		t.Fatalf("operation=%q", got.Operation)
	}

	if _, err = GetRouterTarget(mux, "POST", "/v1/users/42"); err == nil {
		t.Fatal("expected miss for wrong method")
	}
}

func TestGetRouterTarget_DefaultMethodPost(t *testing.T) {
	mux := NewMux()
	if err := mux.routerTree.Add("POST", "/pkg.v1.Svc/Method", "/pkg.v1.Svc/Method", nil); err != nil {
		t.Fatalf("add route: %v", err)
	}

	got, err := GetRouterTarget(mux, "", "/pkg.v1.Svc/Method")
	if err != nil {
		t.Fatalf("GetRouterTarget: %v", err)
	}
	if got.Operation != "/pkg.v1.Svc/Method" {
		t.Fatalf("operation=%q", got.Operation)
	}
}

func TestApplyResponseMetadata_MultiValue(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	applyResponseMetadata(ctx, map[string][]string{
		"x-multi": {"a", "b"},
	})

	vals := ctx.Response().Header.PeekAll("x-multi")
	if len(vals) != 2 {
		t.Fatalf("want 2 header values, got %v", vals)
	}
	if string(vals[0]) != "a" || string(vals[1]) != "b" {
		t.Fatalf("header values=%q %q", vals[0], vals[1])
	}
}

func TestHTTPStatusFromCode_Table(t *testing.T) {
	cases := map[codes.Code]int{
		codes.OK:                200,
		codes.InvalidArgument:   400,
		codes.NotFound:          404,
		codes.Unauthenticated:   401,
		codes.PermissionDenied:  403,
		codes.Unimplemented:     501,
		codes.Unavailable:       503,
		codes.ResourceExhausted: 429,
	}
	for code, want := range cases {
		if got := HTTPStatusFromCode(code); got != want {
			t.Fatalf("code %v: got %d want %d", code, got, want)
		}
	}
}

func TestHTTPFrontend_ServerStreamGRPCWebHeaderMetadataReachesClient(t *testing.T) {
	mux := NewMux()

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	const fullMethod = "/test.v1.StreamService/Watch"
	backend := &fakeClientStream{
		header: metadata.Pairs("x-stream", "live"),
		frames: []proto.Message{
			&structpb.Struct{Fields: map[string]*structpb.Value{
				"msg": structpb.NewStringValue("hello"),
			}},
		},
	}
	method := &methodWrapper{
		srv:            &serviceWrapper{opts: mux.opts, remoteProxyCli: &fakeClientConn{stream: backend}},
		grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true},
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

	req := httptest.NewRequest(fiber.MethodPost, fullMethod, bytes.NewReader([]byte{0, 0, 0, 0, 0}))
	req.Header.Set("Content-Type", "application/grpc-web+proto")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	msgs, trailer := parseGRPCWebResponse(t, body)
	if len(msgs) != 1 {
		t.Fatalf("want 1 streamed message, got %d (body=%q trailer=%v)", len(msgs), body, trailer)
	}
	if got := trailer.Get("x-stream"); got != "live" {
		t.Fatalf("backend header metadata lost: trailer=%v httpHeaders=%v body=%q",
			trailer, resp.Header, body)
	}
	if got := trailer.Get("grpc-status"); got != "0" {
		t.Fatalf("grpc-status=%q want 0, trailer=%v", got, trailer)
	}
}

// newServerStreamMux registers a server-streaming method backed by `backend`.
func newServerStreamMux(t *testing.T, backend Backend) *fiber.App {
	t.Helper()

	mux := NewMux()
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}
	outType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Struct")
	if err != nil {
		t.Fatalf("find output type: %v", err)
	}

	const fullMethod = "/test.v1.StreamService/Watch"
	method := &methodWrapper{
		srv:            &serviceWrapper{opts: mux.opts, remoteProxyCli: backend},
		grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true},
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
	return app
}

func TestHTTPFrontend_ServerStreamNDJSONEmitsOneLinePerMessage(t *testing.T) {
	app := newServerStreamMux(t, &fakeClientConn{stream: &fakeClientStream{
		header: metadata.Pairs("x-stream", "live"),
		frames: []proto.Message{
			&structpb.Struct{Fields: map[string]*structpb.Value{"n": structpb.NewStringValue("1")}},
			&structpb.Struct{Fields: map[string]*structpb.Value{"n": structpb.NewStringValue("2")}},
		},
	}})

	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.StreamService/Watch", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/x-ndjson") {
		t.Fatalf("content-type=%q", got)
	}
	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 NDJSON lines, got %d: %q", len(lines), body)
	}
	for i, line := range lines {
		var obj map[string]any
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("line %d is not a JSON object: %v (%q)", i+1, err, line)
		}
		if want := strconv.Itoa(i + 1); obj["n"] != want {
			t.Fatalf("line %d: n=%v want %q", i+1, obj["n"], want)
		}
	}
}

func TestHTTPFrontend_ServerStreamNDJSONErrorStaysHTTPOk(t *testing.T) {
	// invokeFnConn.NewStream fails, so Dispatch errors after the response head has
	// been committed: the gRPC code can only travel in the payload.
	app := newServerStreamMux(t, &invokeFnConn{})

	req := httptest.NewRequest(fiber.MethodPost, "/test.v1.StreamService/Watch", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200 (already on the wire), body=%q", resp.StatusCode, body)
	}

	var payload struct {
		Code    uint32 `json:"code"`
		Message string `json:"message"`
	}
	if err = json.Unmarshal(bytes.TrimSpace(body), &payload); err != nil {
		t.Fatalf("want a JSON error line, got %q (%v)", body, err)
	}
	if codes.Code(payload.Code) != codes.Unimplemented {
		t.Fatalf("code=%d want Unimplemented, payload=%+v", payload.Code, payload)
	}
}
