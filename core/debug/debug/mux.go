package debug

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/running"
)

var (
	passwd    = running.InstanceID
	once      sync.Once
	configErr error
	startTime = time.Now()
)

// 允许访问的本地 IP 列表
var localHosts = map[string]bool{
	"localhost": true,
	"127.0.0.1": true,
	"::1":       true,
}

// secureCompare 使用常量时间比较防止时序攻击
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// isLocalHost 检查是否为本地访问
func isLocalHost(host string) bool {
	host = strings.Split(host, ":")[0]
	return localHosts[host]
}

// loadConfig 加载配置（只执行一次）
func loadConfig() {
	once.Do(func() {
		configPath := config.GetConfigPath()
		if configPath == "" {
			log.Warn().Msg("debug: config path is empty, using default password")
			return
		}

		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			configErr = errors.WrapCaller(err)
			log.Err(err).Str("path", configPath).Msg("debug: failed to read config file")
			return
		}

		var cfg debug.Config
		if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
			configErr = errors.WrapCaller(err)
			log.Err(err).Msg("debug: failed to parse config file")
			return
		}

		if cfg.Debug.Password != "" {
			passwd = cfg.Debug.Password
		}
	})
}

func init() {
	debug.App().Use(func(c fiber.Ctx) (gErr error) {
		defer recovery.Recovery(func(err error) {
			err = errors.WrapTags(err, errors.Tags{
				"headers": c.GetReqHeaders(),
				"url":     c.Request().URI().String(),
			})
			gErr = c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error":   err.Error(),
				"success": false,
			})
		})

		// 检查是否为 WebSocket 升级请求
		isWebSocket := strings.EqualFold(string(c.Request().Header.Peek("Upgrade")), "websocket")

		// 提取 token
		token := strutil.FirstFnNotEmpty(
			func() string { return c.Query("token") },
			func() string { return string(c.Request().Header.Peek("token")) },
			func() string { return string(c.Request().Header.Peek("Authorization")) },
			func() string { return c.Cookies("token") },
		)

		// 移除 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		// 加载配置
		loadConfig()

		// 非本地访问需要验证 token
		if !isLocalHost(c.Hostname()) {
			if !secureCompare(token, passwd) {
				log.Warn().
					Str("ip", c.IP()).
					Str("path", c.Path()).
					Msg("debug: unauthorized access attempt")

				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized: invalid or missing token",
					"success": false,
				})
			}
		}

		// WebSocket 请求不设置额外的响应头，避免干扰握手
		if !isWebSocket {
			// 设置 cookie 以便后续请求
			if token != "" {
				cc := fasthttp.AcquireCookie()
				defer fasthttp.ReleaseCookie(cc)

				cc.SetKey("token")
				cc.SetValue(token)
				cc.SetHTTPOnly(true)
				cc.SetSameSite(fasthttp.CookieSameSiteStrictMode)
				c.Response().Header.SetCookie(cc)
			}

			// 添加响应头
			c.Set("X-Debug-Version", "1.0")
			c.Set("X-Request-ID", running.InstanceID)
		}

		log.Debug().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Bool("websocket", isWebSocket).
			Msg("debug request")

		return c.Next()
	})
}
