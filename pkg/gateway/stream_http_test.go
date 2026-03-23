package gateway

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
)

func TestStreamHTTP_SendHeader_Idempotent(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	s := &streamHTTP{handler: ctx}

	if err := s.SendHeader(metadata.Pairs("x-test", "v1")); err != nil {
		t.Fatalf("first SendHeader returned error: %v", err)
	}
	if err := s.SendHeader(metadata.Pairs("x-test", "v2")); err != nil {
		t.Fatalf("second SendHeader should be idempotent, got error: %v", err)
	}

	if got := string(ctx.Response().Header.Peek("x-test")); got != "v2" {
		t.Fatalf("header value mismatch: got=%q want=%q", got, "v2")
	}
}

func TestStreamHTTP_SetHeader_AfterSendHeader_NoError(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	s := &streamHTTP{handler: ctx}

	if err := s.SendHeader(metadata.Pairs("x-test", "v1")); err != nil {
		t.Fatalf("SendHeader returned error: %v", err)
	}
	if err := s.SetHeader(metadata.Pairs("x-extra", "ok")); err != nil {
		t.Fatalf("SetHeader after SendHeader should not fail, got error: %v", err)
	}

	if got := string(ctx.Response().Header.Peek("x-extra")); got != "ok" {
		t.Fatalf("header value mismatch: got=%q want=%q", got, "ok")
	}
}
