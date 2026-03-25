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

func App() *fiber.App                           { return app }
func WrapFunc(h http.HandlerFunc) fiber.Handler { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler         { return adaptor.HTTPHandler(h) }
func Get(path string, handler any, handlers ...any) {
	app.Get(path, handler, handlers...)
}

func Head(path string, handler any, handlers ...any) {
	app.Head(path, handler, handlers...)
}

func Post(path string, handler any, handlers ...any) {
	app.Post(path, handler, handlers...)
}

func Put(path string, handler any, handlers ...any) {
	app.Put(path, handler, handlers...)
}

func Delete(path string, handler any, handlers ...any) {
	app.Delete(path, handler, handlers...)
}

func Patch(path string, handler any, handlers ...any) {
	app.Patch(path, handler, handlers...)
}

func Static(prefix, root string, config ...static.Config) {
	app.Use(prefix, static.New(root, config...))
}

func All(path string, handler any, handlers ...any) {
	app.All(path, handler, handlers...)
}

func Group(prefix string, handlers ...any) {
	app.Group(prefix, handlers...)
}

func Route(prefix string, fn func(router fiber.Router), name ...string) {
	app.Route(prefix, fn, name...)
}
