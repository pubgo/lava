package debug

import (
	"fmt"
	"html/template"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var debugStartTime = time.Now()

func init() {
	initDebug()
}

func initDebug() {
	// 主页 - 仪表盘
	debug.Get("/", func(ctx *fiber.Ctx) error {
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

		var routesHTML template.HTML
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
			routesHTML += ui.Card(fmt.Sprintf("📁 %s (%d)", group, len(routes)), tableHTML)
		}

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
</div>`)

		content := statsHTML + `<div class="my-6"></div>` + actionsHTML + routesHTML

		html, _ := ui.Render(ui.PageData{
			Title:       "Debug 控制台",
			Description: "应用调试与监控中心",
			Content:     template.HTML(content),
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
