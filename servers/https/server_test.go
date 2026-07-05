package https

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/log"
	tally "github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/servers/serverhttp"
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

type stubRouter struct {
	prefix string
	mw     []lava.Middleware
}

func (s stubRouter) Middlewares() []lava.Middleware { return s.mw }
func (s stubRouter) Prefix() string                 { return s.prefix }
func (s stubRouter) Router(router fiber.Router) {
	router.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
}

func TestNewServiceName(t *testing.T) {
	t.Parallel()

	svc := New(Params{
		Cfg: &Config{},
		Log: log.GetLogger("test"),
		M:   tally.NoopScope,
	})
	if svc.String() != "http-server" {
		t.Fatalf("String() = %q, want http-server", svc.String())
	}
}

func TestHandlerHttpMiddleInvokesMiddleware(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mw := &testHTTPMiddleware{}
	app.Get("/ping", serverhttp.HandlerMiddleware([]lava.Middleware{mw}))

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

func TestStubRouterRegistersRoute(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	stub := stubRouter{prefix: "/v1"}
	stub.Router(app.Group(stub.Prefix()))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}
