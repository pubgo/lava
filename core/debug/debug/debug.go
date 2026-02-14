package debug

import (
	"fmt"
	"html/template"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var debugStartTime = time.Now()

func init() {
	initDebug()
}

func initDebug() {
	// 主页 - 仪表盘
	debug.Get("/", func(ctx fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// 统计卡片
		statsHTML := ui.Grid(4, ui.StatsCard("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()), "")+
			ui.StatsCard("堆内存", ui.FormatBytes(m.HeapAlloc), fmt.Sprintf("总分配: %s", ui.FormatBytes(m.TotalAlloc)))+
			ui.StatsCard("GC 次数", fmt.Sprintf("%d", m.NumGC), fmt.Sprintf("暂停: %.2fms", float64(m.PauseTotalNs)/1e6))+
			ui.StatsCard("运行时间", formatDuration(time.Since(debugStartTime)), ""))

		// 路由列表
		pathMap := make(map[string][]string)
		stack := ctx.App().Stack()
		for mi := range stack {
			for ri := range stack[mi] {
				route := stack[mi][ri]
				method := route.Method
				if method == "HEAD" {
					continue
				}
				pathMap[route.Path] = append(pathMap[route.Path], method)
			}
		}

		pathList := make([]string, 0, len(pathMap))
		for k := range pathMap {
			pathList = append(pathList, k)
		}
		sort.Strings(pathList)

		// 分组路由
		groups := make(map[string][]routeInfo)
		for _, path := range pathList {
			methods := pathMap[path]
			group := getGroup(path)
			groups[group] = append(groups[group], routeInfo{Path: path, Methods: methods})
		}

		groupNames := make([]string, 0, len(groups))
		for g := range groups {
			groupNames = append(groupNames, g)
		}
		sort.Strings(groupNames)

		// 构建折叠面板的路由列表
		var routeItems string
		for _, group := range groupNames {
			routes := groups[group]
			var rows template.HTML
			for _, r := range routes {
				methodBadges := ""
				for _, m := range r.Methods {
					color := methodColor(m)
					methodBadges += string(ui.Badge(m, color)) + " "
				}
				rows += ui.TR(methodBadges, fmt.Sprintf(`<a href="%s" class="text-blue-400 hover:underline">%s</a>`, r.Path, r.Path))
			}
			tableHTML := ui.Table([]string{"方法", "路径"}) + rows + ui.TableEnd()

			routeItems += fmt.Sprintf(`
                <details class="group border-b border-gray-700 last:border-b-0">
                    <summary class="flex items-center justify-between cursor-pointer px-4 py-3 hover:bg-gray-700/50 transition-colors">
                        <div class="flex items-center space-x-3">
                            <span class="text-gray-400">📁</span>
                            <span class="font-medium text-white">%s</span>
                            <span class="px-2 py-0.5 rounded-full text-xs bg-gray-700 text-gray-300">%d</span>
                        </div>
                        <svg class="w-4 h-4 text-gray-400 transform group-open:rotate-180 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                        </svg>
                    </summary>
                    <div class="px-4 pb-4 bg-gray-800/50">%s</div>
                </details>`, group, len(routes), tableHTML)
		}

		totalRoutes := len(pathList)
		routesHTML := template.HTML(fmt.Sprintf(`
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden mb-4">
    <div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
        <div class="flex items-center space-x-3">
            <h3 class="font-semibold text-white">🛣️ API 路由</h3>
            <span class="px-2 py-0.5 rounded-full text-xs bg-blue-500/20 text-blue-400">%d 个路由</span>
            <span class="px-2 py-0.5 rounded-full text-xs bg-gray-700 text-gray-300">%d 个分组</span>
        </div>
        <div x-data="{ expanded: false }">
            <button @click="expanded = !expanded; document.querySelectorAll('.routes-panel details').forEach(d => d.open = expanded)"
                class="text-xs text-gray-400 hover:text-white transition-colors px-2 py-1 rounded hover:bg-gray-700">
                <span x-text="expanded ? '全部收起' : '全部展开'">全部展开</span>
            </button>
        </div>
    </div>
    <div class="routes-panel max-h-[500px] overflow-y-auto">%s</div>
</div>`, totalRoutes, len(groupNames), routeItems))

		// 快捷操作
		actionsHTML := template.HTML(`
<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
    <a href="/debug/sys/summary" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">💻</div>
        <div class="font-medium">系统信息</div>
        <div class="text-sm text-gray-400">CPU、内存、磁盘</div>
    </a>
    <a href="/debug/runtime/info" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">⚡</div>
        <div class="font-medium">运行时</div>
        <div class="text-sm text-gray-400">Go 运行时指标</div>
    </a>
    <a href="/debug/goroutine/count" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">🔄</div>
        <div class="font-medium">Goroutine</div>
        <div class="text-sm text-gray-400">协程监控</div>
    </a>
    <a href="/debug/supervisor/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">🎛️</div>
        <div class="font-medium">Supervisor</div>
        <div class="text-sm text-gray-400">服务监控与管理</div>
    </a>
    <a href="/debug/scheduler/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">⏰</div>
        <div class="font-medium">Scheduler</div>
        <div class="text-sm text-gray-400">任务调度管理</div>
    </a>
    <a href="/debug/logs/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">📋</div>
        <div class="font-medium">日志查看器</div>
        <div class="text-sm text-gray-400">查询和分析日志</div>
    </a>
    <a href="/debug/health" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">❤️</div>
        <div class="font-medium">健康检查</div>
        <div class="text-sm text-gray-400">组件状态</div>
    </a>
    <a href="/debug/pprof/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">📊</div>
        <div class="font-medium">PProf</div>
        <div class="text-sm text-gray-400">性能分析</div>
    </a>
    <a href="/debug/statsviz/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">📈</div>
        <div class="font-medium">Statsviz</div>
        <div class="text-sm text-gray-400">实时图表</div>
    </a>
    <a href="/debug/log/level" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">📝</div>
        <div class="font-medium">日志级别</div>
        <div class="text-sm text-gray-400">动态调整</div>
    </a>
    <a href="/debug/config" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">⚙️</div>
        <div class="font-medium">配置</div>
        <div class="text-sm text-gray-400">查看配置</div>
    </a>
    <a href="/debug/vars/" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">📊</div>
        <div class="font-medium">Expvar</div>
        <div class="text-sm text-gray-400">运行时变量</div>
    </a>
    <a href="/debug/features" class="bg-gray-800 rounded-lg border border-gray-700 p-4 hover:bg-gray-700 transition-colors">
        <div class="text-2xl mb-2">🚩</div>
        <div class="font-medium">Feature Flags</div>
        <div class="text-sm text-gray-400">功能开关</div>
    </a>
</div>`)

		content := statsHTML + `<div class="my-6"></div>` + actionsHTML + routesHTML

		html, _ := ui.Render(ui.PageData{
			Title:       "Debug 控制台",
			Description: "应用调试与监控中心",
			Content:     content,
		})
		ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
		return ctx.SendString(html)
	})
}

type routeInfo struct {
	Path    string
	Methods []string
}

func getGroup(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		return "/" + parts[0] + "/" + parts[1]
	}
	if len(parts) >= 1 {
		return "/" + parts[0]
	}
	return "/"
}

func methodColor(m string) string {
	colors := map[string]string{
		"GET":    "green",
		"POST":   "blue",
		"PUT":    "yellow",
		"DELETE": "red",
		"PATCH":  "purple",
	}
	if c, ok := colors[m]; ok {
		return c
	}
	return "gray"
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}
