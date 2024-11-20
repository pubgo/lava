package debug

import (
	"net/http"

	_ "github.com/fasthttp/router"
	fiber "github.com/gofiber/fiber/v3"
	adaptor "github.com/gofiber/fiber/v3/middleware/adaptor"
)

type Config struct {
	Debug struct {
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

var app = fiber.New()

func Handler(ctx fiber.Ctx) error {
	app.Handler()(ctx.Context())
	return nil
}

func App() *fiber.App                           { return app }
func WrapFunc(h http.HandlerFunc) fiber.Handler { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler         { return adaptor.HTTPHandler(h) }
func Get(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Get(path, handlers, middleware...)
}

func Head(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Head(path, handlers, middleware...)
}

func Post(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Post(path, handlers, middleware...)
}

func Put(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Put(path, handlers, middleware...)
}

func Delete(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Delete(path, handlers, middleware...)
}

func Connect(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Connect(path, handlers, middleware...)
}

func Options(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Options(path, handlers, middleware...)
}

func Trace(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Trace(path, handlers, middleware...)
}

func Patch(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.Patch(path, handlers, middleware...)
}

func All(path string, handlers fiber.Handler, middleware ...fiber.Handler) {
	app.All(path, handlers, middleware...)
}

func Group(prefix string, handlers ...fiber.Handler) {
	app.Group(prefix, handlers...)
}

func Route(prefix string, fn func(r fiber.Router)) { fn(app.Use(prefix)) }

func Use(args ...any) fiber.Router {
	return app.Use(args...)
}
