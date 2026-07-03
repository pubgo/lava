// Package debug 注册 debug 控制台主页与全局鉴权中间件。
//
// 鉴权策略：
//   - 来自 loopback 地址（127.0.0.1 / ::1）的请求免 token（基于客户端 IP，非 Host 头）
//   - 其他来源须通过 query ?token=、Header token/Authorization 或 Cookie 携带有效 token
//   - 默认 token 启动时随机生成，可通过配置文件 debug.password 覆盖
package debug

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net"
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
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/running"
)

var (
	passwd    string
	once      sync.Once
	configErr error
	startTime = time.Now()
)

// secureCompare 使用常量时间比较防止时序攻击
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// isLocalClient reports whether the request originates from a loopback address.
// Host header is intentionally not used — it is client-controlled and must not
// influence auth decisions.
func isLocalClient(c fiber.Ctx) bool {
	return isLoopbackIP(c.IP())
}

func isLoopbackIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.IsLoopback()
}

func ensurePassword() {
	if passwd != "" {
		return
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		passwd = xid.New().String()
		log.Warn().Err(err).Msg("debug: failed to generate random password, using fallback")
		return
	}
	passwd = hex.EncodeToString(b)
	log.Warn().
		Str("env", running.Env.String()).
		Msg("debug: no debug.password configured; set debug.password in config (auto-generated password active)")
}

// loadConfig 加载配置（只执行一次）
func loadConfig() {
	once.Do(func() {
		configPath := config.GetConfigPath()
		if configPath == "" {
			log.Warn().Msg("debug: config path is empty, using auto-generated password")
			ensurePassword()
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
			return
		}
		ensurePassword()
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
		ensurePassword()

		// 非 loopback 访问需要验证 token
		if !isLocalClient(c) {
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

			// 添加响应头（request ID 与鉴权 token 分离，不复用 InstanceID）
			c.Set("X-Debug-Version", "1.0")
			c.Set("X-Request-ID", xid.New().String())
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
