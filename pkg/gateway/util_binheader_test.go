package gateway

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"

	"github.com/pubgo/lava/v2/pkg/grpcutil"
)

func TestNewIncomingContext_BinHeaderDecode(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString([]byte("hello"))
	h := http.Header{}
	h.Set("x-custom-bin", valid)
	h.Add("x-custom-bin", "!!!not-base64!!!")
	h.Set("x-plain", "ok")
	h.Set("content-type", "application/json")
	h.Set("user-agent", "test-agent")

	_, md := newIncomingContext(context.Background(), h)
	if got := md.Get("x-plain"); len(got) != 1 || got[0] != "ok" {
		t.Fatalf("plain=%v", got)
	}
	got := md.Get("x-custom-bin")
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("bin header=%v, want only decoded value", got)
	}
	if md.Get("content-type") != nil {
		t.Fatalf("reserved content-type should be filtered, got %v", md.Get("content-type"))
	}
	if got := md.Get("user-agent"); len(got) != 1 || got[0] != "test-agent" {
		t.Fatalf("whitelisted user-agent=%v", got)
	}
}

// The gateway's own codec choice makes "content-type" authoritative downstream,
// so it is filtered from the metadata — but the type the caller used is the only
// record of which protocol a request arrived on, and the RPC middleware (and with
// it the RPC metrics' proto label) reads it from grpcutil.MdContentType.
func TestNewIncomingContext_KeepsCallerContentTypeAsMetadata(t *testing.T) {
	h := http.Header{}
	h.Set("content-type", "application/grpc-web+proto")

	_, md := newIncomingContext(context.Background(), h)
	if got := md.Get(grpcutil.MdContentType); len(got) != 1 || got[0] != "application/grpc-web+proto" {
		t.Fatalf("%s=%v, want the caller's content type", grpcutil.MdContentType, got)
	}
	if got := md.Get("content-type"); got != nil {
		t.Fatalf("content-type must stay filtered, got %v", got)
	}
}

// The gRPC-Web frontend normalizes "content-type" for the codec and records the
// caller's own type under this key, so an explicit entry has to win rather than
// race with the normalized one over map iteration order.
func TestNewIncomingContext_ExplicitContentTypeMetadataWins(t *testing.T) {
	for range 20 {
		h := http.Header{}
		h.Set("content-type", "application/grpc+proto")
		h.Set(grpcutil.MdContentType, "application/grpc-web-text+proto")

		_, md := newIncomingContext(context.Background(), h)
		if got := md.Get(grpcutil.MdContentType); len(got) != 1 || got[0] != "application/grpc-web-text+proto" {
			t.Fatalf("%s=%v, want the explicit entry", grpcutil.MdContentType, got)
		}
	}
}

func TestApplyResponseMetadata_SkipsReservedAndEncodesBin(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	applyResponseMetadata(ctx, metadata.MD{
		"x-multi":      {"a", "b"},
		"content-type": {"application/grpc"},
		"grpc-status":  {"0"},
		"x-secret-bin": {"raw-bytes"},
	})

	vals := ctx.Response().Header.PeekAll("x-multi")
	if len(vals) != 2 || string(vals[0]) != "a" || string(vals[1]) != "b" {
		t.Fatalf("x-multi=%v", vals)
	}
	if got := string(ctx.Response().Header.Peek("grpc-status")); got != "" {
		t.Fatalf("grpc-status should be skipped, got %q", got)
	}
	// Fiber may set a default Content-Type; only reject the metadata value we tried to copy.
	if got := string(ctx.Response().Header.Peek("Content-Type")); got == "application/grpc" {
		t.Fatalf("metadata content-type should be skipped, got %q", got)
	}
	bin := string(ctx.Response().Header.Peek("x-secret-bin"))
	want := encodeBinHeader([]byte("raw-bytes"))
	if bin != want {
		t.Fatalf("bin header=%q want %q", bin, want)
	}
}
