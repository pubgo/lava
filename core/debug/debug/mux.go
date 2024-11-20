package debug

import (
	"net/http"
	"os"
	"strings"
	"sync"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/lava/core/debug"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"
)

var (
	passwd = running.InstanceID
	once   sync.Once
)

func init() {
	log.Info().Str("password", running.InstanceID).Msg("debug password")
	debug.App().Use(func(c fiber.Ctx) (gErr error) {
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
			var cfg debug.Config
			assert.Must(yaml.Unmarshal(configBytes, &cfg))
			passwd = cfg.Debug.Password
		})

		host := strings.Split(c.Hostname(), ":")[0]
		if host != "localhost" && host != "127.0.0.1" {
			if token != passwd {
				err := errors.New("token 不存在或者密码不对")
				if ret := result.Of(c.WriteString(err.Error())); ret.IsErr() {
					return errors.WrapCaller(ret.Err())
				}

				if err := c.SendStatus(http.StatusInternalServerError); err != nil {
					return errors.WrapCaller(err)
				}
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
