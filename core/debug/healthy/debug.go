package healthy

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	jjson "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/try"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
	"github.com/pubgo/lava/v2/core/healthy"
)

func init() {
	debug.Get("/health", func(ctx *fiber.Ctx) error {
		dt := make(map[string]*health)
		allHealthy := true
		healthyCount := 0
		unhealthyCount := 0

		for _, name := range healthy.List() {
			h := &health{
				Status: "healthy",
			}
			h.Error = try.Try(func() error {
				defer func(s time.Time) { h.Cost = time.Since(s).String() }(time.Now())
				return healthy.Get(name)(ctx)
			})

			if h.Error != nil {
				h.Status = "unhealthy"
				h.ErrMsg = h.Error.Error()
				h.Error = nil
				allHealthy = false
				unhealthyCount++
			} else {
				healthyCount++
			}
			dt[name] = h
		}

		// JSON 响应
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			bts, err := jjson.Marshal(fiber.Map{
				"status":     statusString(allHealthy),
				"timestamp":  time.Now().Format(time.RFC3339),
				"components": dt,
			})
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				_, err = ctx.Write([]byte(err.Error()))
				return err
			}

			ctx.Response().Header.Set("content-type", "application/json")
			if allHealthy {
				ctx.Status(http.StatusOK)
			} else {
				ctx.Status(http.StatusServiceUnavailable)
			}
			_, err = ctx.Write(bts)
			return err
		}

		// HTML 页面
		overallStatus := "healthy"
		overallColor := "green"
		if !allHealthy {
			overallStatus = "unhealthy"
			overallColor = "red"
		}

		// 统计卡片
		statsContent := ui.StatsCard("总体状态", overallStatus, "")
		statsContent += ui.StatsCard("健康组件", fmt.Sprintf("%d", healthyCount), "")
		statsContent += ui.StatsCard("异常组件", fmt.Sprintf("%d", unhealthyCount), "")
		statsContent += ui.StatsCard("总组件数", fmt.Sprintf("%d", len(dt)), "")

		// 组件列表
		componentsContent := template.HTML("")
		for name, h := range dt {
			statusColor := "green"
			statusIcon := "✓"
			if h.Status == "unhealthy" {
				statusColor = "red"
				statusIcon = "✗"
			}

			errorSection := ""
			if h.ErrMsg != "" {
				errorSection = fmt.Sprintf(`<div class="mt-2 text-sm text-red-400 bg-red-500/10 p-2 rounded">%s</div>`, h.ErrMsg)
			}

			componentsContent += template.HTML(fmt.Sprintf(`
<div class="p-4 bg-gray-700/50 rounded-lg border border-gray-600">
    <div class="flex items-center justify-between mb-2">
        <span class="font-medium text-white">%s</span>
        <span class="px-2 py-1 rounded text-xs font-medium bg-%s-500/20 text-%s-400">%s %s</span>
    </div>
    <div class="text-sm text-gray-400">耗时: %s</div>
    %s
</div>`, name, statusColor, statusColor, statusIcon, h.Status, h.Cost, errorSection))
		}

		if len(dt) == 0 {
			componentsContent = ui.Alert("暂无健康检查组件注册", "yellow")
		}

		// 快速操作
		actionsContent := template.HTML(`
<div class="flex flex-wrap gap-2">
    <a href="/debug/healthz" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-green-600 hover:bg-green-700">存活检查 (Liveness)</a>
    <a href="/debug/readyz" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">就绪检查 (Readiness)</a>
    <button onclick="location.reload()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-gray-600 hover:bg-gray-700">刷新</button>
</div>`)

		// 总体状态显示
		statusBanner := template.HTML(fmt.Sprintf(`
<div class="p-4 rounded-lg border mb-4 bg-%s-500/10 border-%s-500/50">
    <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
            <span class="text-3xl">%s</span>
            <div>
                <div class="font-bold text-%s-400 text-lg">系统%s</div>
                <div class="text-sm text-gray-400">最后检查: %s</div>
            </div>
        </div>
    </div>
</div>`,
			overallColor, overallColor,
			map[bool]string{true: "✅", false: "❌"}[allHealthy],
			overallColor,
			map[bool]string{true: "健康", false: "异常"}[allHealthy],
			time.Now().Format("15:04:05")))

		content := template.HTML(fmt.Sprintf(`
%s
<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">%s</div>
%s
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
    <div class="px-4 py-3 border-b border-gray-700">
        <h3 class="font-semibold text-white">组件状态</h3>
    </div>
    <div class="p-4 grid grid-cols-1 md:grid-cols-2 gap-4">%s</div>
</div>`,
			statusBanner,
			statsContent,
			ui.Card("快速操作", actionsContent),
			componentsContent))

		html, err := ui.Render(ui.PageData{
			Title:       "健康检查",
			Description: "系统健康状态监控",
			Breadcrumb:  []string{"Health"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	// 简单的存活检查
	debug.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	// 就绪检查
	debug.Get("/readyz", func(ctx *fiber.Ctx) error {
		for _, name := range healthy.List() {
			if err := healthy.Get(name)(ctx); err != nil {
				ctx.Status(http.StatusServiceUnavailable)
				return ctx.JSON(fiber.Map{
					"status":    "not ready",
					"component": name,
					"error":     err.Error(),
				})
			}
		}
		return ctx.SendString("ok")
	})
}

func statusString(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

type health struct {
	Status string `json:"status"`
	Cost   string `json:"cost,omitempty"`
	Error  error  `json:"-"`
	ErrMsg string `json:"error,omitempty"`
}
