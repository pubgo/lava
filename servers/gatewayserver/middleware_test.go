package gatewayserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/pkg/lava"
)

type testHTTPMiddleware struct {
	called bool
}

func (m *testHTTPMiddleware) String() string { return "test-http-mw" }

func (m *testHTTPMiddleware) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		m.called = true
		return next(ctx, req)
	}
}

func TestHandlerHttpMiddleRunsMiddleware(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mw := &testHTTPMiddleware{}
	app.Get("/ping", handlerHttpMiddle([]lava.Middleware{mw}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound && resp.StatusCode != fiber.StatusOK {
		// middleware chain runs before route miss; either OK or 404 is fine
		t.Logf("status=%d", resp.StatusCode)
	}
	if !mw.called {
		t.Fatal("middleware was not invoked")
	}
}
