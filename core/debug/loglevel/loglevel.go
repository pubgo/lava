package loglevel

import (
	"fmt"
	"html/template"
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/log"
	"github.com/rs/zerolog"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var currentLevel atomic.Value

func Register() {
	currentLevel.Store(zerolog.GlobalLevel().String())

	debug.Get("/log/level", func(ctx fiber.Ctx) error {
		// JSON 格式响应
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			return ctx.JSON(fiber.Map{
				"level":            currentLevel.Load(),
				"available_levels": getAvailableLevels(),
			})
		}

		// HTML 页面
		current := currentLevel.Load().(string)
		levels := getAvailableLevels()

		// 构建级别选择器
		levelOptions := ""
		for _, level := range levels {
			selected := ""
			if level == current {
				selected = "selected"
			}
			levelColor := getLevelColor(level)
			levelOptions += fmt.Sprintf(`<option value="%s" %s class="%s">%s</option>`, level, selected, levelColor, strings.ToUpper(level))
		}

		// 级别颜色映射
		levelBadges := template.HTML("")
		for _, level := range levels {
			active := ""
			if level == current {
				active = "ring-2 ring-white"
			}
			levelBadges += template.HTML(fmt.Sprintf(`<span class="px-3 py-1.5 rounded text-xs font-medium %s %s cursor-pointer" onclick="setLevel('%s')">%s</span>`,
				getLevelBgColor(level), active, level, strings.ToUpper(level)))
		}

		content := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
    %s
    %s
    %s
</div>
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden mb-4">
    <div class="px-4 py-3 border-b border-gray-700">
        <h3 class="font-semibold text-white">日志级别选择</h3>
    </div>
    <div class="p-4">
        <div class="flex flex-wrap gap-2 mb-4">%s</div>
        <div class="mt-4 p-3 bg-gray-700/50 rounded">
            <div class="text-sm text-gray-400 mb-2">级别说明:</div>
            <ul class="text-sm text-gray-300 space-y-1">
                <li><span class="text-purple-400">TRACE</span> - 最详细的追踪信息</li>
                <li><span class="text-blue-400">DEBUG</span> - 调试信息</li>
                <li><span class="text-green-400">INFO</span> - 一般信息</li>
                <li><span class="text-yellow-400">WARN</span> - 警告信息</li>
                <li><span class="text-red-400">ERROR</span> - 错误信息</li>
                <li><span class="text-red-600">FATAL</span> - 致命错误</li>
                <li><span class="text-red-800">PANIC</span> - 严重错误</li>
                <li><span class="text-gray-500">DISABLED</span> - 禁用日志</li>
            </ul>
        </div>
    </div>
</div>
<div id="result" class="hidden p-4 rounded-lg mb-4"></div>
<script>
async function setLevel(level) {
    const resultDiv = document.getElementById('result');
    try {
        const res = await fetch('/debug/log/level?level=' + level, { method: 'POST' });
        const data = await res.json();
        if (data.success) {
            resultDiv.className = 'p-4 rounded-lg mb-4 bg-green-500/10 border border-green-500/50 text-green-400';
            resultDiv.innerHTML = '✓ 日志级别已从 <strong>' + data.old_level.toUpperCase() + '</strong> 更改为 <strong>' + data.new_level.toUpperCase() + '</strong>';
            resultDiv.classList.remove('hidden');
            setTimeout(() => location.reload(), 1000);
        } else {
            throw new Error(data.error || '设置失败');
        }
    } catch (e) {
        resultDiv.className = 'p-4 rounded-lg mb-4 bg-red-500/10 border border-red-500/50 text-red-400';
        resultDiv.innerHTML = '✗ 设置失败: ' + e.message;
        resultDiv.classList.remove('hidden');
    }
}
</script>`,
			ui.StatsCard("当前级别", strings.ToUpper(current), ""),
			ui.StatsCard("可用级别", fmt.Sprintf("%d", len(levels)), ""),
			ui.StatsCard("全局日志", "zerolog", ""),
			levelBadges))

		html, err := ui.Render(ui.PageData{
			Title:       "日志级别管理",
			Description: "动态调整应用日志级别",
			Breadcrumb:  []string{"Log", "Level"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	debug.Put("/log/level", func(ctx fiber.Ctx) error {
		return setLogLevel(ctx)
	})

	debug.Post("/log/level", func(ctx fiber.Ctx) error {
		return setLogLevel(ctx)
	})
}

func setLogLevel(ctx fiber.Ctx) error {
	type request struct {
		Level string `json:"level" form:"level" query:"level"`
	}

	var req request
	if err := ctx.Bind().Body(&req); err != nil {
		req.Level = ctx.Query("level")
	}

	if req.Level == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":            "level is required",
			"available_levels": getAvailableLevels(),
		})
	}

	level, err := zerolog.ParseLevel(strings.ToLower(req.Level))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":            "invalid log level: " + req.Level,
			"available_levels": getAvailableLevels(),
		})
	}

	oldLevel := currentLevel.Load()
	zerolog.SetGlobalLevel(level)
	currentLevel.Store(level.String())

	log.Info().
		Str("old_level", oldLevel.(string)).
		Str("new_level", level.String()).
		Msg("log level changed")

	return ctx.JSON(fiber.Map{
		"success":   true,
		"old_level": oldLevel,
		"new_level": level.String(),
	})
}

func getAvailableLevels() []string {
	return []string{
		zerolog.TraceLevel.String(),
		zerolog.DebugLevel.String(),
		zerolog.InfoLevel.String(),
		zerolog.WarnLevel.String(),
		zerolog.ErrorLevel.String(),
		zerolog.FatalLevel.String(),
		zerolog.PanicLevel.String(),
		zerolog.Disabled.String(),
	}
}

func getLevelColor(level string) string {
	colors := map[string]string{
		"trace":    "text-purple-400",
		"debug":    "text-blue-400",
		"info":     "text-green-400",
		"warn":     "text-yellow-400",
		"error":    "text-red-400",
		"fatal":    "text-red-600",
		"panic":    "text-red-800",
		"disabled": "text-gray-500",
	}
	if c, ok := colors[level]; ok {
		return c
	}
	return "text-gray-400"
}

func getLevelBgColor(level string) string {
	colors := map[string]string{
		"trace":    "bg-purple-500/20 text-purple-400",
		"debug":    "bg-blue-500/20 text-blue-400",
		"info":     "bg-green-500/20 text-green-400",
		"warn":     "bg-yellow-500/20 text-yellow-400",
		"error":    "bg-red-500/20 text-red-400",
		"fatal":    "bg-red-600/20 text-red-500",
		"panic":    "bg-red-800/20 text-red-600",
		"disabled": "bg-gray-500/20 text-gray-500",
	}
	if c, ok := colors[level]; ok {
		return c
	}
	return "bg-gray-500/20 text-gray-400"
}
