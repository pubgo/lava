package debug

import (
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

var app = fiber.New()

func App() *fiber.App { return app }

func WrapFunc(h http.HandlerFunc) fiber.Handler                { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler                        { return adaptor.HTTPHandler(h) }
func Get(path string, handlers ...fiber.Handler) fiber.Router  { return app.Get(path, handlers...) }
func Head(path string, handlers ...fiber.Handler) fiber.Router { return app.Head(path, handlers...) }
func Post(path string, handlers ...fiber.Handler) fiber.Router { return app.Post(path, handlers...) }
func Put(path string, handlers ...fiber.Handler) fiber.Router  { return app.Put(path, handlers...) }
func Delete(path string, handlers ...fiber.Handler) fiber.Router {
	return app.Delete(path, handlers...)
}

func Connect(path string, handlers ...fiber.Handler) fiber.Router {
	return app.Connect(path, handlers...)
}

func Options(path string, handlers ...fiber.Handler) fiber.Router {
	return app.Options(path, handlers...)
}

func Trace(path string, handlers ...fiber.Handler) fiber.Router { return app.Trace(path, handlers...) }
func Patch(path string, handlers ...fiber.Handler) fiber.Router { return app.Patch(path, handlers...) }
func Static(prefix, root string, config ...fiber.Static) fiber.Router {
	return app.Static(prefix, root, config...)
}

func All(path string, handlers ...fiber.Handler) fiber.Router { return app.All(path, handlers...) }
func Group(prefix string, handlers ...fiber.Handler) fiber.Router {
	return app.Group(prefix, handlers...)
}

func Route(prefix string, fn func(router fiber.Router), name ...string) fiber.Router {
	return app.Route(prefix, fn, name...)
}
