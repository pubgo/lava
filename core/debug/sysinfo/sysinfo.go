package sysinfo

import (
	"fmt"
	"html/template"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/running"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	psprocess "github.com/shirou/gopsutil/v3/process"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var sysStartTime = time.Now()

func init() {
	// 系统信息仪表板 HTML 页面
	debug.Get("/sys", func(ctx *fiber.Ctx) error {
		vmem, _ := mem.VirtualMemory()
		cpuPercent, _ := cpu.Percent(0, false)
		loadAvg, _ := load.Avg()
		hostInfo, _ := host.Info()

		// JSON 响应
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			pid := int32(os.Getpid())
			proc, _ := psprocess.NewProcess(pid)
			procCPU, _ := proc.CPUPercent()
			procMem, _ := proc.MemoryPercent()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return ctx.JSON(fiber.Map{
				"timestamp": time.Now().Format(time.RFC3339),
				"system": fiber.Map{
					"hostname":       hostInfo.Hostname,
					"platform":       hostInfo.Platform,
					"cpu_percent":    cpuPercent,
					"load_avg":       loadAvg,
					"memory_total":   vmem.Total,
					"memory_used":    vmem.Used,
					"memory_percent": vmem.UsedPercent,
				},
				"process": fiber.Map{
					"pid":            pid,
					"cpu_percent":    procCPU,
					"memory_percent": procMem,
					"goroutines":     runtime.NumGoroutine(),
					"heap_alloc":     m.HeapAlloc,
				},
			})
		}

		// HTML 页面
		partitions, _ := disk.Partitions(false)

		pid := int32(os.Getpid())
		proc, _ := psprocess.NewProcess(pid)
		procCPU, _ := proc.CPUPercent()
		procMem, _ := proc.MemoryPercent()
		numThreads, _ := proc.NumThreads()
		numFDs, _ := proc.NumFDs()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// CPU 使用率
		cpuVal := float64(0)
		if len(cpuPercent) > 0 {
			cpuVal = cpuPercent[0]
		}
		cpuColor := "green"
		if cpuVal > 80 {
			cpuColor = "red"
		} else if cpuVal > 60 {
			cpuColor = "yellow"
		}

		// 内存使用率
		memColor := "green"
		if vmem.UsedPercent > 80 {
			memColor = "red"
		} else if vmem.UsedPercent > 60 {
			memColor = "yellow"
		}

		// 系统概览卡片
		statsContent := ui.StatsCard("主机名", hostInfo.Hostname, hostInfo.Platform+" "+hostInfo.PlatformVersion)
		statsContent += ui.StatsCard("CPU", fmt.Sprintf("%.1f%%", cpuVal), fmt.Sprintf("%d 核心, 负载: %.2f", runtime.NumCPU(), loadAvg.Load1))
		statsContent += ui.StatsCard("内存", fmt.Sprintf("%.1f%%", vmem.UsedPercent), fmt.Sprintf("%s / %s", ui.FormatBytes(vmem.Used), ui.FormatBytes(vmem.Total)))
		statsContent += ui.StatsCard("运行时间", time.Since(sysStartTime).Round(time.Second).String(), fmt.Sprintf("开机: %s", formatDuration(hostInfo.Uptime)))

		// CPU 详情
		cpuContent := template.HTML(fmt.Sprintf(`
<div class="space-y-4">
    <div class="flex justify-between items-center">
        <span class="text-gray-400">CPU 使用率</span>
        <span class="font-mono text-white">%.1f%%</span>
    </div>
    %s
    <div class="grid grid-cols-3 gap-4 mt-4">
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">1 分钟负载</div>
            <div class="font-bold text-white">%.2f</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">5 分钟负载</div>
            <div class="font-bold text-white">%.2f</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">15 分钟负载</div>
            <div class="font-bold text-white">%.2f</div>
        </div>
    </div>
</div>`, cpuVal, ui.ProgressBar(cpuVal, cpuColor), loadAvg.Load1, loadAvg.Load5, loadAvg.Load15))

		// 内存详情
		memContent := template.HTML(fmt.Sprintf(`
<div class="space-y-4">
    <div class="flex justify-between items-center">
        <span class="text-gray-400">内存使用率</span>
        <span class="font-mono text-white">%s / %s (%.1f%%)</span>
    </div>
    %s
    <div class="grid grid-cols-2 gap-4 mt-4">
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">可用内存</div>
            <div class="font-bold text-white">%s</div>
        </div>
        <div class="text-center p-3 bg-gray-700/50 rounded">
            <div class="text-xs text-gray-400">缓冲/缓存</div>
            <div class="font-bold text-white">%s</div>
        </div>
    </div>
</div>`, ui.FormatBytes(vmem.Used), ui.FormatBytes(vmem.Total), vmem.UsedPercent,
			ui.ProgressBar(vmem.UsedPercent, memColor),
			ui.FormatBytes(vmem.Available), ui.FormatBytes(vmem.Cached)))

		// 进程信息
		processContent := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">PID</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">CPU</div>
        <div class="font-bold text-white">%.1f%%</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">内存</div>
        <div class="font-bold text-white">%.1f%%</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">Goroutines</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">线程</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">文件描述符</div>
        <div class="font-bold text-white">%d</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">堆内存</div>
        <div class="font-bold text-white">%s</div>
    </div>
    <div class="text-center p-3 bg-gray-700/50 rounded">
        <div class="text-xs text-gray-400">GC 次数</div>
        <div class="font-bold text-white">%d</div>
    </div>
</div>`, pid, procCPU, procMem, runtime.NumGoroutine(), numThreads, numFDs, ui.FormatBytes(m.HeapAlloc), m.NumGC))

		// 磁盘信息
		diskContent := template.HTML("")
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			diskColor := "green"
			if usage.UsedPercent > 90 {
				diskColor = "red"
			} else if usage.UsedPercent > 70 {
				diskColor = "yellow"
			}
			diskContent += template.HTML(fmt.Sprintf(`
<div class="p-3 bg-gray-700/50 rounded mb-2">
    <div class="flex justify-between items-center mb-2">
        <span class="font-medium text-white">%s</span>
        <span class="text-sm text-gray-400">%s</span>
    </div>
    <div class="text-sm text-gray-400 mb-2">%s / %s (%.1f%%)</div>
    %s
</div>`, p.Mountpoint, p.Fstype, ui.FormatBytes(usage.Used), ui.FormatBytes(usage.Total), usage.UsedPercent, ui.ProgressBar(usage.UsedPercent, diskColor)))
		}
		if diskContent == "" {
			diskContent = template.HTML(`<div class="text-gray-500">无磁盘信息</div>`)
		}

		// 快速操作
		actionsContent := template.HTML(`
<div class="flex flex-wrap gap-2">
    <a href="/debug/sys/cpu" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">CPU 详情</a>
    <a href="/debug/sys/memory" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-green-600 hover:bg-green-700">内存详情</a>
    <a href="/debug/sys/disk" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-purple-600 hover:bg-purple-700">磁盘详情</a>
    <a href="/debug/sys/network" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-yellow-600 hover:bg-yellow-700">网络详情</a>
    <a href="/debug/sys/process" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-red-600 hover:bg-red-700">进程详情</a>
</div>`)

		// 组合内容
		content := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">%s</div>
%s
<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    %s
    %s
</div>
%s
%s`,
			statsContent,
			ui.Card("快速操作", actionsContent),
			ui.Card("CPU 状态", cpuContent),
			ui.Card("内存状态", memContent),
			ui.Card("当前进程", processContent),
			ui.Card("磁盘使用", diskContent)))

		html, err := ui.Render(ui.PageData{
			Title:       "系统信息",
			Description: "系统资源监控和进程信息",
			Breadcrumb:  []string{"System"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	debug.Get("/sys/info", func(ctx *fiber.Ctx) error {
		hostInfo, _ := host.Info()
		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"host":      hostInfo,
			"go": fiber.Map{
				"version":       runtime.Version(),
				"os":            runtime.GOOS,
				"arch":          runtime.GOARCH,
				"num_cpu":       runtime.NumCPU(),
				"gomaxprocs":    runtime.GOMAXPROCS(0),
				"num_goroutine": runtime.NumGoroutine(),
			},
			"app": fiber.Map{
				"project":     running.Project(),
				"version":     running.Version(),
				"hostname":    running.Hostname,
				"instance_id": running.InstanceID,
				"uptime":      time.Since(sysStartTime).String(),
			},
		})
	})

	debug.Get("/sys/cpu", func(ctx *fiber.Ctx) error {
		cpuInfo, _ := cpu.Info()
		cpuPercent, _ := cpu.Percent(time.Second, false)
		cpuTimes, _ := cpu.Times(false)
		loadAvg, _ := load.Avg()

		return ctx.JSON(fiber.Map{
			"timestamp":  time.Now().Format(time.RFC3339),
			"info":       cpuInfo,
			"percent":    cpuPercent,
			"times":      cpuTimes,
			"load_avg":   loadAvg,
			"num_cpu":    runtime.NumCPU(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
		})
	})

	debug.Get("/sys/memory", func(ctx *fiber.Ctx) error {
		vmem, _ := mem.VirtualMemory()
		swap, _ := mem.SwapMemory()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"virtual":   vmem,
			"swap":      swap,
			"go_heap": fiber.Map{
				"alloc_bytes": m.HeapAlloc,
				"sys_bytes":   m.HeapSys,
				"idle_bytes":  m.HeapIdle,
				"inuse_bytes": m.HeapInuse,
				"objects":     m.HeapObjects,
			},
		})
	})

	debug.Get("/sys/disk", func(ctx *fiber.Ctx) error {
		partitions, _ := disk.Partitions(false)
		var diskUsages []fiber.Map
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err == nil {
				diskUsages = append(diskUsages, fiber.Map{
					"device":     p.Device,
					"mountpoint": p.Mountpoint,
					"fstype":     p.Fstype,
					"total":      usage.Total,
					"used":       usage.Used,
					"free":       usage.Free,
					"percent":    usage.UsedPercent,
				})
			}
		}

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"disks":     diskUsages,
		})
	})

	debug.Get("/sys/network", func(ctx *fiber.Ctx) error {
		interfaces, _ := net.Interfaces()
		ioCounters, _ := net.IOCounters(true)
		connections, _ := net.Connections("all")

		return ctx.JSON(fiber.Map{
			"timestamp":        time.Now().Format(time.RFC3339),
			"interfaces":       interfaces,
			"io_counters":      ioCounters,
			"connection_count": len(connections),
		})
	})

	debug.Get("/sys/process", func(ctx *fiber.Ctx) error {
		pid := int32(os.Getpid())
		proc, err := psprocess.NewProcess(pid)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		cpuPercent, _ := proc.CPUPercent()
		memInfo, _ := proc.MemoryInfo()
		memPercent, _ := proc.MemoryPercent()
		numThreads, _ := proc.NumThreads()
		numFDs, _ := proc.NumFDs()
		ioCounters, _ := proc.IOCounters()
		connections, _ := proc.Connections()
		openFiles, _ := proc.OpenFiles()
		createTime, _ := proc.CreateTime()

		return ctx.JSON(fiber.Map{
			"timestamp":      time.Now().Format(time.RFC3339),
			"pid":            pid,
			"cpu_percent":    cpuPercent,
			"memory_info":    memInfo,
			"memory_percent": memPercent,
			"num_threads":    numThreads,
			"num_fds":        numFDs,
			"io_counters":    ioCounters,
			"connections":    len(connections),
			"open_files":     len(openFiles),
			"create_time":    time.UnixMilli(createTime).Format(time.RFC3339),
			"uptime":         time.Since(sysStartTime).String(),
		})
	})

	debug.Get("/sys/summary", func(ctx *fiber.Ctx) error {
		vmem, _ := mem.VirtualMemory()
		cpuPercent, _ := cpu.Percent(0, false)
		loadAvg, _ := load.Avg()
		hostInfo, _ := host.Info()

		pid := int32(os.Getpid())
		proc, _ := psprocess.NewProcess(pid)
		procCPU, _ := proc.CPUPercent()
		procMem, _ := proc.MemoryPercent()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		return ctx.JSON(fiber.Map{
			"timestamp": time.Now().Format(time.RFC3339),
			"system": fiber.Map{
				"hostname":       hostInfo.Hostname,
				"os":             hostInfo.OS,
				"platform":       hostInfo.Platform,
				"cpu_percent":    cpuPercent,
				"load_avg":       loadAvg,
				"memory_total":   vmem.Total,
				"memory_used":    vmem.Used,
				"memory_percent": vmem.UsedPercent,
			},
			"process": fiber.Map{
				"pid":            pid,
				"cpu_percent":    procCPU,
				"memory_percent": procMem,
				"goroutines":     runtime.NumGoroutine(),
				"heap_alloc":     m.HeapAlloc,
				"uptime":         time.Since(sysStartTime).String(),
			},
		})
	})
}

func formatDuration(seconds uint64) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
