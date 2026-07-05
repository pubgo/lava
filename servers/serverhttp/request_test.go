package serverhttp_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/servers/serverhttp"
)

func TestRequestOperation(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/ping", func(ctx fiber.Ctx) error {
		req := serverhttp.NewRequest(ctx).(*serverhttp.Request)
		if req.Kind() != lava.RequestKindHttp {
			t.Fatalf("kind=%q", req.Kind())
		}
		if req.Operation() == "" {
			t.Fatal("empty operation")
		}
		return ctx.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestRequestResponseImplementsLava(t *testing.T) {
	t.Parallel()
	var _ lava.Request = (*serverhttp.Request)(nil)
	var _ lava.Response = (*serverhttp.Response)(nil)
}

func TestResponseStreamMatchesFiber(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/ok", func(ctx fiber.Ctx) error {
		rsp := serverhttp.NewResponse(ctx).(*serverhttp.Response)
		if rsp.Stream() != ctx.Response().IsBodyStream() {
			t.Fatalf("stream mismatch: %v vs %v", rsp.Stream(), ctx.Response().IsBodyStream())
		}
		return ctx.SendString("ok")
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ok", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}
