package debug

import (
	"net/http"

	_ "github.com/fasthttp/router"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	Debug struct {
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

var app = fiber.New()

func Handler(ctx *fiber.Ctx) error {
	app.Handler()(ctx.Context())
	return nil
}

func App() *fiber.App                                    { return app }
func WrapFunc(h http.HandlerFunc) fiber.Handler          { return adaptor.HTTPHandlerFunc(h) }
func Wrap(h http.Handler) fiber.Handler                  { return adaptor.HTTPHandler(h) }
func Get(path string, handlers ...fiber.Handler)         { app.Get(path, handlers...) }
func Head(path string, handlers ...fiber.Handler)        { app.Head(path, handlers...) }
func Post(path string, handlers ...fiber.Handler)        { app.Post(path, handlers...) }
func Put(path string, handlers ...fiber.Handler)         { app.Put(path, handlers...) }
func Delete(path string, handlers ...fiber.Handler)      { app.Delete(path, handlers...) }
func Connect(path string, handlers ...fiber.Handler)     { app.Connect(path, handlers...) }
func Options(path string, handlers ...fiber.Handler)     { app.Options(path, handlers...) }
func Trace(path string, handlers ...fiber.Handler)       { app.Trace(path, handlers...) }
func Patch(path string, handlers ...fiber.Handler)       { app.Patch(path, handlers...) }
func Static(prefix, root string, config ...fiber.Static) { app.Static(prefix, root, config...) }
func All(path string, handlers ...fiber.Handler)         { app.All(path, handlers...) }
func Group(prefix string, handlers ...fiber.Handler)     { app.Group(prefix, handlers...) }
func Route(prefix string, fn func(router fiber.Router), name ...string) {
	app.Route(prefix, fn, name...)
}
