package debug

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type Config struct {
	Debug struct {
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

var app = fiber.New()

func handlersToAny(handlers ...fiber.Handler) []any {
	if len(handlers) == 0 {
		return nil
	}
	args := make([]any, len(handlers))
	for i, h := range handlers {
		args[i] = h
	}
	return args
}

func App() *fiber.App                           { return app }
func WrapFunc(h http.HandlerFunc) fiber.Handler { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler         { return adaptor.HTTPHandler(h) }
func Get(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Get(path, handler, handlersToAny(handlers...)...)
}

func Head(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Head(path, handler, handlersToAny(handlers...)...)
}

func Post(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Post(path, handler, handlersToAny(handlers...)...)
}

func Put(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Put(path, handler, handlersToAny(handlers...)...)
}

func Delete(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Delete(path, handler, handlersToAny(handlers...)...)
}

func Connect(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Connect(path, handler, handlersToAny(handlers...)...)
}

func Options(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Options(path, handler, handlersToAny(handlers...)...)
}

func Trace(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Trace(path, handler, handlersToAny(handlers...)...)
}

func Patch(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.Patch(path, handler, handlersToAny(handlers...)...)
}

func Static(prefix, root string, config ...static.Config) {
	app.Use(prefix, static.New(root, config...))
}

func All(path string, handler fiber.Handler, handlers ...fiber.Handler) {
	app.All(path, handler, handlersToAny(handlers...)...)
}

func Group(prefix string, handlers ...fiber.Handler) {
	app.Group(prefix, handlersToAny(handlers...)...)
}

func Route(prefix string, fn func(router fiber.Router), name ...string) {
	app.Route(prefix, fn, name...)
}
