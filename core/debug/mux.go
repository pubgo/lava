// Package debug 提供全局 debug HTTP 路由聚合器（基于 Fiber）。
//
// 各子模块通过 debug/bootstrap.RegisterAll() 向全局 App 注册路由；
// 主 HTTP/gRPC 服务器通过 app.Use("/debug", debug.App()) 挂载。
//
// 鉴权由 debug/debug 子包的全局中间件负责：
//   - 来自 loopback 地址（127.0.0.1 / ::1）的请求免 token（基于客户端 IP）
//   - 非本地访问需携带 token（query/header/cookie），可在配置 debug.password 中设置
package debug

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Config 是 debug 模块的 YAML 配置结构。
type Config struct {
	Debug struct {
		// Password 是非本地访问 debug 端点所需的 token。
		Password string `yaml:"password"`
	} `yaml:"debug"`
}

// app 是全局 debug Fiber 实例，所有 debug 子模块共享。
var app = fiber.New()

// App 返回全局 debug Fiber 实例，供主服务器挂载。
func App() *fiber.App { return app }

// WrapFunc 将标准库 http.HandlerFunc 适配为 Fiber Handler。
func WrapFunc(h http.HandlerFunc) fiber.Handler { return adaptor.HTTPHandlerFunc(h) }

// Wrap 将标准库 http.Handler 适配为 Fiber Handler。
func Wrap(h http.Handler) fiber.Handler { return adaptor.HTTPHandler(h) }

// Get 在全局 debug App 上注册 GET 路由。
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
