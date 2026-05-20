package runtime

import (
	"fmt"
	"html/template"
	"runtime"
	rd "runtime/debug"
	"runtime/metrics"
	"sort"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var runtimeStartTime = time.Now()

func init() {
	// Runtime 仪表板 HTML 页面
	debug.Get("/runtime", func(ctx fiber.Ctx) error {
		// 如果请求 JSON 格式
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return ctx.JSON(fiber.Map{
				"timestamp":     time.Now().Format(time.RFC3339),
				"go_version":    runtime.Version(),
				"go_os":         runtime.GOOS,
				"go_arch":       runtime.GOARCH,
				"num_cpu":       runtime.NumCPU(),
				"num_goroutine": runtime.NumGoroutine(),
				"gomaxprocs":    runtime.GOMAXPROCS(0),
				"heap_alloc":    m.HeapAlloc,
				"heap_sys":      m.HeapSys,
				"uptime":        time.Since(runtimeStartTime).String(),
			})
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		heapPercent := float64(m.HeapInuse) / float64(m.HeapSys) * 100
		heapColor := "green"
		if heapPercent > 80 {
			heapColor = "red"
		} else if heapPercent > 60 {
			heapColor = "yellow"
		}

		// 构建统计卡片
		statsContent := ui.StatsCard("Go 版本", runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
		statsContent += ui.StatsCard("CPU", fmt.Sprintf("%d/%d", runtime.GOMAXPROCS(0), runtime.NumCPU()), "GOMAXPROCS/NumCPU")
		statsContent += ui.StatsCard("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()), "当前活跃")
		statsContent += ui.StatsCard("Cgo 调用", fmt.Sprintf("%d", runtime.NumCgoCall()), "累计")

		// 内存卡片内容
		memContent := template.HTML(fmt.Sprintf(`
<div class="space-y-4">
    <div class="flex justify-between items-center">
        <span class="text-gray-400">堆内存使用</span>
        <span class="font-mono text-white">%s / %s</span>
    </div>
    %s
    <div class="grid grid-cols-2 gap-4 mt-4">
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">Heap Alloc</div>
            <div class="font-bold text-white">%s</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">Heap Objects</div>
            <div class="font-bold text-white">%d</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">Stack Inuse</div>
            <div class="font-bold text-white">%s</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">Total Alloc</div>
            <div class="font-bold text-white">%s</div>
        </div>
    </div>
</div>`, ui.FormatBytes(m.HeapInuse), ui.FormatBytes(m.HeapSys), ui.ProgressBar(heapPercent, heapColor),
			ui.FormatBytes(m.HeapAlloc), m.HeapObjects, ui.FormatBytes(m.StackInuse), ui.FormatBytes(m.TotalAlloc)))

		// GC 卡片内容
		gcContent := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-2 gap-4">
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">GC 次数</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">GC CPU</div>
        <div class="font-bold text-white">%.4f%%</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">上次 GC</div>
        <div class="font-bold text-white">%s</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">下次 GC 阈值</div>
        <div class="font-bold text-white">%s</div>
    </div>
</div>`, m.NumGC, m.GCCPUFraction*100, time.Unix(0, int64(m.LastGC)).Format("15:04:05"), ui.FormatBytes(m.NextGC)))

		// 操作按钮
		actionsContent := template.HTML(`
<div class="flex flex-wrap gap-2">
    <a href="/debug/runtime/memory" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">内存详情</a>
    <a href="/debug/runtime/gc" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-green-600 hover:bg-green-700">GC 详情</a>
    <a href="/debug/runtime/metrics" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-purple-600 hover:bg-purple-700">所有指标</a>
    <button onclick="triggerGC()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-yellow-600 hover:bg-yellow-700">触发 GC</button>
    <button onclick="freeMemory()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700">释放内存</button>
</div>
<script>
async function triggerGC() {
    if (!confirm('确定要触发 GC 吗？')) return;
    try {
        const res = await fetch('/debug/runtime/gc/trigger', { method: 'POST' });
        const data = await res.json();
        alert('GC 完成！释放了 ' + formatBytes(data.freed_bytes));
        location.reload();
    } catch (e) {
        alert('GC 失败: ' + e.message);
    }
}
async function freeMemory() {
    if (!confirm('确定要释放内存给操作系统吗？')) return;
    try {
        const res = await fetch('/debug/runtime/freemem', { method: 'POST' });
        const data = await res.json();
        alert('内存释放完成！');
        location.reload();
    } catch (e) {
        alert('释放失败: ' + e.message);
    }
}
function formatBytes(b) {
    if (b < 1024) return b + ' B';
    if (b < 1048576) return (b/1024).toFixed(2) + ' KB';
    if (b < 1073741824) return (b/1048576).toFixed(2) + ' MB';
    return (b/1073741824).toFixed(2) + ' GB';
}
</script>`)

		// 组合内容
		content := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">%s</div>
%s
<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    %s
    %s
</div>
<div class="mt-6 bg-gray-800 rounded-lg border border-gray-700 p-4">
    <h3 class="font-semibold text-white mb-4">运行时 %s</h3>
    <div class="text-gray-400 text-sm">Sys: %s | Uptime: %s</div>
</div>`,
			statsContent,
			ui.CardWithAction("快捷操作", "", actionsContent),
			ui.Card("内存状态", memContent),
			ui.Card("GC 状态", gcContent),
			time.Since(runtimeStartTime).Round(time.Second).String(),
			ui.FormatBytes(m.Sys),
			time.Since(runtimeStartTime).Round(time.Second).String()))

		html, err := ui.Render(ui.PageData{
			Title:       "Runtime 监控",
			Description: "Go 运行时状态和内存监控",
			Breadcrumb:  []string{"Runtime"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	// 获取所有 runtime/metrics 指标
	debug.Get("/runtime/metrics", func(ctx fiber.Ctx) error {
		descs := metrics.All()
		samples := make([]metrics.Sample, len(descs))
		for i := range descs {
			samples[i].Name = descs[i].Name
		}
		metrics.Read(samples)

		result := make(map[string]any, len(samples))
		for _, sample := range samples {
			result[sample.Name] = formatMetricValue(sample.Value)
		}

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"metrics":   result,
		})
	})

	// 内存统计
	debug.Get("/runtime/memory", func(ctx fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"memory": fiber.Map{
				"heap_alloc_bytes":    m.HeapAlloc,
				"heap_sys_bytes":      m.HeapSys,
				"heap_idle_bytes":     m.HeapIdle,
				"heap_inuse_bytes":    m.HeapInuse,
				"heap_released_bytes": m.HeapReleased,
				"heap_objects":        m.HeapObjects,
				"stack_inuse_bytes":   m.StackInuse,
				"stack_sys_bytes":     m.StackSys,
				"alloc_bytes":         m.Alloc,
				"total_alloc_bytes":   m.TotalAlloc,
				"sys_bytes":           m.Sys,
				"mallocs":             m.Mallocs,
				"frees":               m.Frees,
				"gc_sys_bytes":        m.GCSys,
				"gc_next_bytes":       m.NextGC,
				"gc_num":              m.NumGC,
				"gc_cpu_fraction":     m.GCCPUFraction,
			},
		})
	})

	// 运行时信息
	debug.Get("/runtime/info", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"timestamp":      time.Now().Format(time.RFC3339),
			"go_version":     runtime.Version(),
			"go_os":          runtime.GOOS,
			"go_arch":        runtime.GOARCH,
			"num_cpu":        runtime.NumCPU(),
			"num_goroutine":  runtime.NumGoroutine(),
			"num_cgo_call":   runtime.NumCgoCall(),
			"gomaxprocs":     runtime.GOMAXPROCS(0),
			"uptime":         time.Since(runtimeStartTime).String(),
			"uptime_seconds": time.Since(runtimeStartTime).Seconds(),
		})
	})

	// GC 统计
	debug.Get("/runtime/gc", func(ctx fiber.Ctx) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"gc": fiber.Map{
				"num_gc":          m.NumGC,
				"num_forced_gc":   m.NumForcedGC,
				"gc_cpu_fraction": m.GCCPUFraction,
				"pause_total_ns":  m.PauseTotalNs,
				"pause_total_ms":  float64(m.PauseTotalNs) / 1e6,
				"last_gc":         time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
				"next_gc_bytes":   m.NextGC,
				"recent_pauses":   recentPauses(&m),
			},
		})
	})

	// 手动触发 GC
	debug.Post("/runtime/gc/trigger", func(ctx fiber.Ctx) error {
		before := runtime.NumGoroutine()
		var mBefore runtime.MemStats
		runtime.ReadMemStats(&mBefore)

		runtime.GC()

		var mAfter runtime.MemStats
		runtime.ReadMemStats(&mAfter)
		after := runtime.NumGoroutine()

		return ctx.JSON(fiber.Map{
			"success":   true,
			"timestamp": time.Now().Format(time.RFC3339),
			"before": fiber.Map{
				"heap_alloc_bytes": mBefore.HeapAlloc,
				"heap_objects":     mBefore.HeapObjects,
				"goroutines":       before,
			},
			"after": fiber.Map{
				"heap_alloc_bytes": mAfter.HeapAlloc,
				"heap_objects":     mAfter.HeapObjects,
				"goroutines":       after,
			},
			"freed_bytes": int64(mBefore.HeapAlloc) - int64(mAfter.HeapAlloc),
		})
	})

	// 释放内存给操作系统
	debug.Post("/runtime/freemem", func(ctx fiber.Ctx) error {
		var mBefore runtime.MemStats
		runtime.ReadMemStats(&mBefore)

		runtime.GC()
		rd.FreeOSMemory()

		var mAfter runtime.MemStats
		runtime.ReadMemStats(&mAfter)

		return ctx.JSON(fiber.Map{
			"success":   true,
			"timestamp": time.Now().Format(time.RFC3339),
			"before": fiber.Map{
				"heap_sys_bytes":      mBefore.HeapSys,
				"heap_released_bytes": mBefore.HeapReleased,
			},
			"after": fiber.Map{
				"heap_sys_bytes":      mAfter.HeapSys,
				"heap_released_bytes": mAfter.HeapReleased,
			},
		})
	})

	// 可用的 metrics 描述
	debug.Get("/runtime/metrics/desc", func(ctx fiber.Ctx) error {
		descs := metrics.All()
		result := make([]fiber.Map, 0, len(descs))

		for _, desc := range descs {
			result = append(result, fiber.Map{
				"name":        desc.Name,
				"description": desc.Description,
				"kind":        kindString(desc.Kind),
				"cumulative":  desc.Cumulative,
			})
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i]["name"].(string) < result[j]["name"].(string)
		})

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"count":     len(result),
			"metrics":   result,
		})
	})
}

func formatMetricValue(v metrics.Value) any {
	switch v.Kind() {
	case metrics.KindUint64:
		return v.Uint64()
	case metrics.KindFloat64:
		return v.Float64()
	case metrics.KindFloat64Histogram:
		h := v.Float64Histogram()
		return fiber.Map{
			"buckets": h.Buckets,
			"counts":  h.Counts,
		}
	case metrics.KindBad:
		return "bad metric"
	default:
		return "unknown"
	}
}

func kindString(k metrics.ValueKind) string {
	switch k {
	case metrics.KindUint64:
		return "uint64"
	case metrics.KindFloat64:
		return "float64"
	case metrics.KindFloat64Histogram:
		return "float64_histogram"
	case metrics.KindBad:
		return "bad"
	default:
		return "unknown"
	}
}

func recentPauses(m *runtime.MemStats) []uint64 {
	n := int(m.NumGC)
	if n > 256 {
		n = 256
	}
	if n > 10 {
		n = 10
	}

	pauses := make([]uint64, n)
	for i := 0; i < n; i++ {
		idx := (int(m.NumGC) - 1 - i) % 256
		pauses[i] = m.PauseNs[idx]
	}
	return pauses
}
