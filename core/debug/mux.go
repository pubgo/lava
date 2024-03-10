package debug

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"net/http"
)

type Config struct {
	Debug struct {
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

var app = fiber.New()

func App() *fiber.App                                    { return app }
func WrapFunc(h http.HandlerFunc) fiber.Handler          { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler                  { return adaptor.HTTPHandler(h) }
func Get(path string, handler fiber.Handler)             { app.Get(path, handler) }
func Head(path string, handler fiber.Handler)            { app.Head(path, handler) }
func Post(path string, handler fiber.Handler)            { app.Post(path, handler) }
func Put(path string, handler fiber.Handler)             { app.Put(path, handler) }
func Delete(path string, handler fiber.Handler)          { app.Delete(path, handler) }
func Connect(path string, handler fiber.Handler)         { app.Connect(path, handler) }
func Options(path string, handler fiber.Handler)         { app.Options(path, handler) }
func Trace(path string, handler fiber.Handler)           { app.Trace(path, handler) }
func Patch(path string, handler fiber.Handler)           { app.Patch(path, handler) }
func Static(prefix, root string, config ...fiber.Static) { app.Static(prefix, root, config...) }
func All(path string, handler fiber.Handler)             { app.All(path, handler) }
func Group(prefix string, handler fiber.Handler)         { app.Group(prefix, handler) }
func Route(prefix string, fn func(router fiber.Router)) {
	fn(app.Group(prefix))
}
