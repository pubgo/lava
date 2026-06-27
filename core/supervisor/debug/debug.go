// Package debug 注册 Supervisor 的 debug UI 与 REST 控制 API。
//
// 路由挂载在 /debug/supervisor 下，鉴权由 core/debug 全局中间件统一处理
//（非 localhost 访问需 token，详见 core/debug 包文档）。
package debug

import (
	_ "embed"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/supervisor"
)

// Register 将 Supervisor 管理页面与控制 API 注册到 debug 路由组。
func Register(mgr *supervisor.Manager) {
	h := &handler{mgr: mgr}

	debug.Route("/supervisor", func(router fiber.Router) {
		// 主页面 - UI 界面
		router.Get("/", h.handleDebugPage)

		// API 端点
		router.Get("/api/services", h.handleAPIServices)
		router.Get("/api/service/:name", h.handleAPIServiceDetail)
		router.Post("/api/service/:name/restart", h.handleAPIRestartService)
		router.Post("/api/service/:name/stop", h.handleAPIStopService)
		router.Post("/api/service/:name/start", h.handleAPIStartService)
		router.Post("/api/service/:name/reset", h.handleAPIResetService)
		router.Post("/api/services/restart", h.handleAPIRestartAll)

		// 兼容旧端点
		router.Get("services", h.handleAPIServices)
	})
}

type handler struct {
	mgr *supervisor.Manager
}

func (h *handler) handleAPIServices(ctx fiber.Ctx) error {
	services := h.mgr.GetServicesInfo()
	return ctx.JSON(services)
}

func (h *handler) handleAPIServiceDetail(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	info, err := h.mgr.GetServiceInfo(name)
	if err != nil {
		return ctx.Status(statusCodeFromErr(err)).JSON(fiber.Map{
			"error": "service not found",
			"name":  name,
		})
	}
	return ctx.JSON(info)
}

func (h *handler) handleAPIRestartService(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	if err := h.mgr.RestartService(name); err != nil {
		return ctx.Status(statusCodeFromErr(err)).JSON(fiber.Map{
			"error": err.Error(),
			"name":  name,
		})
	}
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "service restarted",
		"name":    name,
	})
}

func (h *handler) handleAPIStopService(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	if err := h.mgr.StopService(name); err != nil {
		return ctx.Status(statusCodeFromErr(err)).JSON(fiber.Map{
			"error": err.Error(),
			"name":  name,
		})
	}
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "service stopped",
		"name":    name,
	})
}

func (h *handler) handleAPIStartService(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	if err := h.mgr.StartService(name); err != nil {
		return ctx.Status(statusCodeFromErr(err)).JSON(fiber.Map{
			"error": err.Error(),
			"name":  name,
		})
	}
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "service started",
		"name":    name,
	})
}

func (h *handler) handleAPIResetService(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	if err := h.mgr.ResetService(name); err != nil {
		return ctx.Status(statusCodeFromErr(err)).JSON(fiber.Map{
			"error": err.Error(),
			"name":  name,
		})
	}
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "service reset",
		"name":    name,
	})
}

func (h *handler) handleAPIRestartAll(ctx fiber.Ctx) error {
	if err := h.mgr.RestartServices(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "all services restarted",
	})
}

func (h *handler) handleDebugPage(ctx fiber.Ctx) error {
	html := supervisorDebugPageHTML
	ctx.Set("Content-Type", "text/html; charset=utf-8")
	return ctx.SendString(html)
}

//go:embed index.html
var supervisorDebugPageHTML string

func statusCodeFromErr(err error) int {
	if errors.Is(err, supervisor.ErrServiceNotFound) {
		return fiber.StatusNotFound
	}

	if errors.Is(err, supervisor.ErrServiceAlreadyExists) {
		return fiber.StatusConflict
	}

	return fiber.StatusInternalServerError
}
