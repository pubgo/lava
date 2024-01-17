package debug

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/strutil"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Config struct {
	Debug struct {
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

var app = fiber.New()
var passwd string = running.InstanceID
var once sync.Once

func init() {
	log.Info().Str("password", running.InstanceID).Msg("debug password")
	app.Use(func(c *fiber.Ctx) (gErr error) {
		defer recovery.Recovery(func(err error) {
			err = errors.WrapTag(err,
				errors.T("headers", c.GetReqHeaders()),
				errors.T("url", c.Request().URI().String()),
			)
			gErr = c.JSON(err)
		})

		token := strutil.FirstFnNotEmpty(
			func() string { return c.Query("token") },
			func() string { return string(c.Request().Header.Peek("token")) },
			func() string { return c.Cookies("token") },
		)

		once.Do(func() {
			configBytes := assert.Must1(os.ReadFile(config.GetConfigPath()))
			var cfg Config
			assert.Must(yaml.Unmarshal(configBytes, &cfg))
			passwd = cfg.Debug.Password
		})

		host := strings.Split(c.Hostname(), ":")[0]
		if host != "localhost" && host != "127.0.0.1" {
			if token != passwd {
				var err = errors.New("token 不存在或者密码不对")
				c.WriteString(err.Error())
				c.SendStatus(http.StatusInternalServerError)
				return err
			}
		}

		cc := fasthttp.AcquireCookie()
		defer fasthttp.ReleaseCookie(cc)

		cc.SetKey("token")
		cc.SetValue(token)
		c.Response().Header.SetCookie(cc)

		log.Info().Str("path", c.Request().URI().String()).Msg("request")

		return c.Next()
	})
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
