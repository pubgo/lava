package goroutine

import (
	"fmt"
	"html/template"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var (
	monitoring     atomic.Bool
	monitorMu      sync.Mutex
	goroutineStats []goroutineStat
	maxStats       = 100
)

type goroutineStat struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

func init() {
	// Goroutine 仪表板 HTML 页面
	debug.Get("/goroutine", func(ctx *fiber.Ctx) error {
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			return ctx.JSON(fiber.Map{
				"timestamp": time.Now().Format(time.RFC3339),
				"count":     runtime.NumGoroutine(),
			})
		}

		monitorMu.Lock()
		stats := make([]goroutineStat, len(goroutineStats))
		copy(stats, goroutineStats)
		monitorMu.Unlock()

		isMonitoring := monitoring.Load()
		monitorStatus := "停止"
		if isMonitoring {
			monitorStatus = "运行中"
		}

		// 构建统计卡片
		statsContent := ui.StatsCard("Goroutine 数量", fmt.Sprintf("%d", runtime.NumGoroutine()), "当前")
		statsContent += ui.StatsCard("监控状态", monitorStatus, fmt.Sprintf("已采集 %d 个样本", len(stats)))
		statsContent += ui.StatsCard("最大样本", fmt.Sprintf("%d", maxStats), "保留数量")
		statsContent += ui.StatsCard("CPU 核心", fmt.Sprintf("%d", runtime.NumCPU()), "GOMAXPROCS: "+fmt.Sprintf("%d", runtime.GOMAXPROCS(0)))

		// 趋势分析
		trendContent := template.HTML("")
		if len(stats) >= 2 {
			first := stats[0].Count
			last := stats[len(stats)-1].Count
			diff := last - first
			trend := "稳定"
			trendColor := "blue"
			if diff > 10 {
				trend = "上升 ↑"
				trendColor = "yellow"
			} else if diff < -10 {
				trend = "下降 ↓"
				trendColor = "green"
			}
			isPossibleLeak := diff > 50
			leakWarning := ""
			if isPossibleLeak {
				leakWarning = fmt.Sprintf(`<div class="mt-2 p-2 bg-red-500/20 border border-red-500/50 rounded text-red-400 text-sm">⚠️ 可能存在 Goroutine 泄漏！增长了 %d 个</div>`, diff)
			}
			trendContent = template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-4 gap-4">
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">初始数量</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">当前数量</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">变化</div>
        <div class="font-bold text-white">%+d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">趋势</div>
        <div class="font-bold">%s</div>
    </div>
</div>
%s`, first, last, diff, ui.Badge(trend, trendColor), leakWarning))
		} else {
			trendContent = ui.Alert("需要更多数据点，请先启动监控", "yellow")
		}

		// 图表数据
		chartData := "[]"
		if len(stats) > 0 {
			points := make([]string, 0, len(stats))
			for _, s := range stats {
				points = append(points, fmt.Sprintf(`{x:"%s",y:%d}`, s.Timestamp.Format("15:04:05"), s.Count))
			}
			chartData = "[" + concatStrings(points, ",") + "]"
		}

		// 操作按钮和图表
		actionsContent := template.HTML(fmt.Sprintf(`
<div class="flex flex-wrap gap-2 mb-4">
    <button onclick="toggleMonitor()" id="monitorBtn" class="px-3 py-1.5 rounded text-sm font-medium text-white %s transition-colors">
        %s
    </button>
    <button onclick="clearStats()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-yellow-600 hover:bg-yellow-700">清除数据</button>
    <a href="/debug/goroutine/stack" target="_blank" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-purple-600 hover:bg-purple-700">查看堆栈</a>
    <a href="/debug/goroutine/profile" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">Goroutine 分析</a>
    <button onclick="checkLeak()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700">泄漏检测</button>
</div>
<div id="chart" class="h-64 bg-gray-900 rounded p-4">
    <canvas id="goroutineChart"></canvas>
</div>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script>
let isMonitoring = %v;
const chartData = %s;
let chart;

document.addEventListener('DOMContentLoaded', function() {
    const ctx = document.getElementById('goroutineChart').getContext('2d');
    chart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: chartData.map(d => d.x),
            datasets: [{
                label: 'Goroutine 数量',
                data: chartData.map(d => d.y),
                borderColor: 'rgb(59, 130, 246)',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                fill: true,
                tension: 0.3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: { beginAtZero: false, grid: { color: 'rgba(255,255,255,0.1)' }, ticks: { color: '#9ca3af' } },
                x: { grid: { color: 'rgba(255,255,255,0.1)' }, ticks: { color: '#9ca3af' } }
            },
            plugins: { legend: { labels: { color: '#fff' } } }
        }
    });
    
    if (isMonitoring) startAutoRefresh();
});

let refreshInterval;
function startAutoRefresh() {
    refreshInterval = setInterval(async () => {
        const res = await fetch('/debug/goroutine/count');
        const data = await res.json();
        const now = new Date().toLocaleTimeString('zh-CN', {hour12: false});
        chart.data.labels.push(now);
        chart.data.datasets[0].data.push(data.count);
        if (chart.data.labels.length > 100) {
            chart.data.labels.shift();
            chart.data.datasets[0].data.shift();
        }
        chart.update('none');
    }, 1000);
}

async function toggleMonitor() {
    const endpoint = isMonitoring ? '/debug/goroutine/monitor/stop' : '/debug/goroutine/monitor/start';
    await fetch(endpoint, { method: 'POST' });
    isMonitoring = !isMonitoring;
    const btn = document.getElementById('monitorBtn');
    if (isMonitoring) {
        btn.textContent = '停止监控';
        btn.className = btn.className.replace('bg-green-600 hover:bg-green-700', 'bg-red-600 hover:bg-red-700');
        startAutoRefresh();
    } else {
        btn.textContent = '开始监控';
        btn.className = btn.className.replace('bg-red-600 hover:bg-red-700', 'bg-green-600 hover:bg-green-700');
        clearInterval(refreshInterval);
    }
}

async function clearStats() {
    if (!confirm('确定要清除监控数据吗？')) return;
    await fetch('/debug/goroutine/monitor/clear', { method: 'POST' });
    location.reload();
}

async function checkLeak() {
    const res = await fetch('/debug/goroutine/leak/check');
    const data = await res.json();
    if (data.status === 'insufficient_data') {
        alert('数据不足，请先启动监控收集数据');
    } else if (data.possible_leak) {
        alert('⚠️ 警告：可能存在 Goroutine 泄漏！\n\n趋势: ' + data.trend + '\n初始: ' + data.initial_count + '\n当前: ' + data.current_count + '\n增长: ' + data.difference);
    } else {
        alert('✅ Goroutine 数量正常\n\n趋势: ' + data.trend + '\n初始: ' + data.initial_count + '\n当前: ' + data.current_count);
    }
}
</script>`,
			map[bool]string{true: "bg-red-600 hover:bg-red-700", false: "bg-green-600 hover:bg-green-700"}[isMonitoring],
			map[bool]string{true: "停止监控", false: "开始监控"}[isMonitoring],
			isMonitoring, chartData))

		// 组合内容
		content := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">%s</div>
%s
%s`,
			statsContent,
			ui.Card("趋势分析", trendContent),
			ui.Card("监控图表", actionsContent)))

		html, err := ui.Render(ui.PageData{
			Title:       "Goroutine 监控",
			Description: "Goroutine 数量监控和泄漏检测",
			Breadcrumb:  []string{"Goroutine"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	debug.Get("/goroutine/count", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"count":     runtime.NumGoroutine(),
		})
	})

	debug.Get("/goroutine/stack", func(ctx *fiber.Ctx) error {
		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, true)
		ctx.Set("Content-Type", "text/plain; charset=utf-8")
		return ctx.Send(buf[:n])
	})

	debug.Get("/goroutine/stats", func(ctx *fiber.Ctx) error {
		monitorMu.Lock()
		stats := make([]goroutineStat, len(goroutineStats))
		copy(stats, goroutineStats)
		monitorMu.Unlock()

		return ctx.JSON(fiber.Map{
			"monitoring": monitoring.Load(),
			"count":      len(stats),
			"stats":      stats,
		})
	})

	debug.Post("/goroutine/monitor/start", func(ctx *fiber.Ctx) error {
		if monitoring.Load() {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "monitoring already started",
			})
		}

		monitoring.Store(true)
		go monitorGoroutines()

		return ctx.JSON(fiber.Map{
			"success": true,
			"message": "monitoring started",
		})
	})

	debug.Post("/goroutine/monitor/stop", func(ctx *fiber.Ctx) error {
		if !monitoring.Load() {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "monitoring not started",
			})
		}

		monitoring.Store(false)

		return ctx.JSON(fiber.Map{
			"success": true,
			"message": "monitoring stopped",
		})
	})

	debug.Post("/goroutine/monitor/clear", func(ctx *fiber.Ctx) error {
		monitorMu.Lock()
		goroutineStats = nil
		monitorMu.Unlock()

		return ctx.JSON(fiber.Map{
			"success": true,
			"message": "stats cleared",
		})
	})

	debug.Get("/goroutine/profile", func(ctx *fiber.Ctx) error {
		profiles := make([]fiber.Map, 0)
		profileMap := make(map[string]int)

		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, true)
		stacks := string(buf[:n])

		lines := splitLines(stacks)
		var currentStack string
		for _, line := range lines {
			if line == "" {
				if currentStack != "" {
					profileMap[currentStack]++
					currentStack = ""
				}
				continue
			}
			if len(line) > 0 && line[0] != '\t' && line[0] != ' ' {
				if currentStack != "" {
					profileMap[currentStack]++
				}
				currentStack = line
			}
		}

		type kv struct {
			Key   string
			Value int
		}
		var sorted []kv
		for k, v := range profileMap {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Value > sorted[j].Value
		})

		for _, item := range sorted {
			profiles = append(profiles, fiber.Map{
				"state": item.Key,
				"count": item.Value,
			})
		}

		return ctx.JSON(fiber.Map{
			"timestamp":       time.Now().Format(time.RFC3339),
			"total":           runtime.NumGoroutine(),
			"unique_profiles": len(profiles),
			"profiles":        profiles,
		})
	})

	debug.Get("/goroutine/leak/check", func(ctx *fiber.Ctx) error {
		monitorMu.Lock()
		stats := make([]goroutineStat, len(goroutineStats))
		copy(stats, goroutineStats)
		monitorMu.Unlock()

		if len(stats) < 2 {
			return ctx.JSON(fiber.Map{
				"status":  "insufficient_data",
				"message": "need more data points, start monitoring first",
			})
		}

		first := stats[0].Count
		last := stats[len(stats)-1].Count
		diff := last - first
		trend := "stable"
		if diff > 10 {
			trend = "increasing"
		} else if diff < -10 {
			trend = "decreasing"
		}

		return ctx.JSON(fiber.Map{
			"status":        "ok",
			"trend":         trend,
			"initial_count": first,
			"current_count": last,
			"difference":    diff,
			"sample_count":  len(stats),
			"possible_leak": trend == "increasing" && diff > 50,
		})
	})
}

func monitorGoroutines() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for monitoring.Load() {
		select {
		case <-ticker.C:
			stat := goroutineStat{
				Timestamp: time.Now(),
				Count:     runtime.NumGoroutine(),
			}

			monitorMu.Lock()
			goroutineStats = append(goroutineStats, stat)
			if len(goroutineStats) > maxStats {
				goroutineStats = goroutineStats[len(goroutineStats)-maxStats:]
			}
			monitorMu.Unlock()

			log.Debug().Int("count", stat.Count).Msg("goroutine count")
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	var current string
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += fmt.Sprintf("%c", c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func concatStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
