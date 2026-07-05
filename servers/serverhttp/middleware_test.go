package serverhttp

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/pkg/lava"
)

type testMiddleware struct {
	called bool
}

func (m *testMiddleware) String() string { return "test-mw" }

func (m *testMiddleware) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		m.called = true
		return next(ctx, req)
	}
}

func TestHandlerMiddlewareRunsChain(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mw := &testMiddleware{}
	app.Get("/ping", HandlerMiddleware([]lava.Middleware{mw}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound && resp.StatusCode != fiber.StatusOK {
		t.Logf("status=%d", resp.StatusCode)
	}
	if !mw.called {
		t.Fatal("middleware was not invoked")
	}
}
