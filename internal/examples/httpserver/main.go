package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/recovery"

	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/logging"
	_ "github.com/pubgo/lava/v2/core/logging/logext/slog"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/servers/https"
)

var _ lava.HttpRouter = (*helloRouter)(nil)

type helloRouter struct{}

func (h *helloRouter) Prefix() string { return "/api" }

func (h *helloRouter) Middlewares() []lava.Middleware { return nil }

func (h *helloRouter) Router(router fiber.Router) {
	router.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
			"service": "http-only",
		})
	})

	router.Get("/hello/:name", func(c fiber.Ctx) error {
		name := c.Params("name")
		if name == "" {
			name = "world"
		}

		return c.SendString(fmt.Sprintf("hello, %s", name))
	})
}

func main() {
	defer recovery.Exit()

	version.SetVersion("v1.0.0")
	version.SetProject("httpserver-example")

	env.Reload()
	env.LoadFiles(".env").Must()

	config.SetConfigPath("internal/examples/httpserver/config.yaml")

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(func() lava.HttpRouter { return &helloRouter{} })

	// 运行纯 HTTP 服务：
	//   go run ./internal/examples/httpserver http
	// 注意：这里没有注册任何 gRPC router，因此使用 http 命令时只会启动 HTTP 服务。
	lavabuilder.Run(builder)
}

type Config struct {
	metrics.MetricConfigLoader   `yaml:",inline"`
	logging.LogConfigLoader      `yaml:",inline"`
	https.HttpServerConfigLoader `yaml:",inline"`
}
