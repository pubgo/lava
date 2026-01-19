package ui

import (
	"bytes"
	"fmt"
	"html/template"
)

// BaseTemplate 基础 HTML 模板
const BaseTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Debug Console</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        [x-cloak] { display: none !important; }
        .loading { animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: .5; } }
        pre { white-space: pre-wrap; word-wrap: break-word; }
    </style>
    {{if .ExtraHead}}{{.ExtraHead}}{{end}}
</head>
<body class="bg-gray-900 text-gray-100 min-h-screen">
    <nav class="bg-gray-800 border-b border-gray-700 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4">
            <div class="flex items-center justify-between h-14">
                <div class="flex items-center space-x-4">
                    <a href="/debug/" class="text-xl font-bold text-blue-400 hover:text-blue-300">🔧 Debug</a>
                    <span class="text-gray-500">|</span>
                    <a href="/debug/sys" class="text-gray-300 hover:text-white text-sm">系统</a>
                    <a href="/debug/runtime" class="text-gray-300 hover:text-white text-sm">运行时</a>
                    <a href="/debug/goroutine" class="text-gray-300 hover:text-white text-sm">Goroutine</a>
                    <a href="/debug/health" class="text-gray-300 hover:text-white text-sm">健康</a>
                    <a href="/debug/log/level" class="text-gray-300 hover:text-white text-sm">日志</a>
                    <a href="/debug/config" class="text-gray-300 hover:text-white text-sm">配置</a>
                    <a href="/debug/pprof/" class="text-gray-300 hover:text-white text-sm">PProf</a>
                </div>
                <div class="flex items-center space-x-2 text-xs text-gray-500" x-data="{ time: '' }" x-init="setInterval(() => time = new Date().toLocaleString('zh-CN'), 1000)">
                    <span x-text="time"></span>
                </div>
            </div>
        </div>
    </nav>
    <main class="max-w-7xl mx-auto px-4 py-6">
        {{if .Breadcrumb}}
        <nav class="text-sm mb-4">
            <ol class="flex items-center space-x-2">
                <li><a href="/debug/" class="text-blue-400 hover:text-blue-300">Debug</a></li>
                {{range .Breadcrumb}}
                <li class="text-gray-500">/</li>
                <li class="text-gray-300">{{.}}</li>
                {{end}}
            </ol>
        </nav>
        {{end}}
        <div class="mb-6">
            <h1 class="text-2xl font-bold text-white">{{.Title}}</h1>
            {{if .Description}}<p class="text-gray-400 mt-1">{{.Description}}</p>{{end}}
        </div>
        {{.Content}}
    </main>
    <footer class="border-t border-gray-800 mt-auto py-4">
        <div class="max-w-7xl mx-auto px-4 text-center text-gray-500 text-sm">
            Debug Console | <a href="/debug/version" class="text-blue-400 hover:underline">Version</a>
        </div>
    </footer>
</body>
</html>`

// PageData 页面数据
type PageData struct {
	Title       string
	Description string
	Breadcrumb  []string
	Content     template.HTML
	ExtraHead   template.HTML
}

// Render 渲染基础模板
func Render(data PageData) (string, error) {
	tmpl, err := template.New("base").Parse(BaseTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Card 渲染卡片
func Card(title string, content template.HTML) template.HTML {
	return template.HTML(fmt.Sprintf(`
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden mb-4">
    <div class="px-4 py-3 border-b border-gray-700">
        <h3 class="font-semibold text-white">%s</h3>
    </div>
    <div class="p-4">%s</div>
</div>`, title, content))
}

// CardWithAction 带操作按钮的卡片
func CardWithAction(title string, actions, content template.HTML) template.HTML {
	return template.HTML(fmt.Sprintf(`
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden mb-4">
    <div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
        <h3 class="font-semibold text-white">%s</h3>
        <div class="flex items-center space-x-2">%s</div>
    </div>
    <div class="p-4">%s</div>
</div>`, title, actions, content))
}

// StatsCard 统计卡片
func StatsCard(label, value, subtext string) template.HTML {
	sub := ""
	if subtext != "" {
		sub = fmt.Sprintf(`<div class="text-xs text-gray-500 mt-1">%s</div>`, subtext)
	}
	return template.HTML(fmt.Sprintf(`
<div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
    <div class="text-gray-400 text-sm">%s</div>
    <div class="text-2xl font-bold text-white mt-1">%s</div>
    %s
</div>`, label, value, sub))
}

// Badge 徽章
func Badge(text, color string) template.HTML {
	colorClass := map[string]string{
		"green":  "bg-green-500/20 text-green-400",
		"red":    "bg-red-500/20 text-red-400",
		"yellow": "bg-yellow-500/20 text-yellow-400",
		"blue":   "bg-blue-500/20 text-blue-400",
		"gray":   "bg-gray-500/20 text-gray-400",
		"purple": "bg-purple-500/20 text-purple-400",
	}
	cls := colorClass[color]
	if cls == "" {
		cls = colorClass["gray"]
	}
	return template.HTML(fmt.Sprintf(`<span class="px-2 py-1 rounded text-xs font-medium %s">%s</span>`, cls, text))
}

// Button 按钮
func Button(text, onclick, color string) template.HTML {
	colorClass := map[string]string{
		"blue":   "bg-blue-600 hover:bg-blue-700",
		"green":  "bg-green-600 hover:bg-green-700",
		"red":    "bg-red-600 hover:bg-red-700",
		"yellow": "bg-yellow-600 hover:bg-yellow-700",
		"gray":   "bg-gray-600 hover:bg-gray-700",
	}
	cls := colorClass[color]
	if cls == "" {
		cls = colorClass["blue"]
	}
	return template.HTML(fmt.Sprintf(`<button onclick="%s" class="px-3 py-1.5 rounded text-sm font-medium text-white %s transition-colors">%s</button>`, onclick, cls, text))
}

// Link 链接按钮
func Link(text, href, color string) template.HTML {
	colorClass := map[string]string{
		"blue":  "text-blue-400 hover:text-blue-300",
		"green": "text-green-400 hover:text-green-300",
		"gray":  "text-gray-400 hover:text-gray-300",
	}
	cls := colorClass[color]
	if cls == "" {
		cls = colorClass["blue"]
	}
	return template.HTML(fmt.Sprintf(`<a href="%s" class="%s text-sm">%s</a>`, href, cls, text))
}

// Table 表格开始
func Table(headers []string) template.HTML {
	var h string
	for _, header := range headers {
		h += fmt.Sprintf(`<th class="px-4 py-2 text-left text-gray-300 font-medium">%s</th>`, header)
	}
	return template.HTML(fmt.Sprintf(`
<div class="overflow-x-auto">
    <table class="w-full text-sm">
        <thead class="bg-gray-700/50"><tr>%s</tr></thead>
        <tbody class="divide-y divide-gray-700">`, h))
}

// TableEnd 表格结束
func TableEnd() template.HTML {
	return template.HTML(`</tbody></table></div>`)
}

// TR 表格行
func TR(cells ...string) template.HTML {
	var c string
	for _, cell := range cells {
		c += fmt.Sprintf(`<td class="px-4 py-2 text-gray-300">%s</td>`, cell)
	}
	return template.HTML(fmt.Sprintf(`<tr class="hover:bg-gray-700/30">%s</tr>`, c))
}

// Grid 网格布局
func Grid(cols int, content template.HTML) template.HTML {
	return template.HTML(fmt.Sprintf(`<div class="grid grid-cols-1 md:grid-cols-%d gap-4">%s</div>`, cols, content))
}

// JSONBlock JSON 代码块
func JSONBlock(id string) template.HTML {
	return template.HTML(fmt.Sprintf(`<pre id="%s" class="bg-gray-900 rounded p-4 text-sm text-gray-300 overflow-auto max-h-96"></pre>`, id))
}

// FormatBytes 格式化字节
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), []string{"KB", "MB", "GB", "TB"}[exp])
}

// FormatPercent 格式化百分比
func FormatPercent(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}

// ProgressBar 进度条
func ProgressBar(percent float64, color string) template.HTML {
	colorClass := map[string]string{
		"green":  "bg-green-500",
		"red":    "bg-red-500",
		"yellow": "bg-yellow-500",
		"blue":   "bg-blue-500",
	}
	cls := colorClass[color]
	if cls == "" {
		cls = colorClass["blue"]
	}
	return template.HTML(fmt.Sprintf(`
<div class="w-full bg-gray-700 rounded-full h-2">
    <div class="%s h-2 rounded-full" style="width: %.1f%%"></div>
</div>`, cls, percent))
}

// Alert 警告框
func Alert(text, color string) template.HTML {
	colorClass := map[string]string{
		"green":  "bg-green-500/10 border-green-500/50 text-green-400",
		"red":    "bg-red-500/10 border-red-500/50 text-red-400",
		"yellow": "bg-yellow-500/10 border-yellow-500/50 text-yellow-400",
		"blue":   "bg-blue-500/10 border-blue-500/50 text-blue-400",
	}
	cls := colorClass[color]
	if cls == "" {
		cls = colorClass["blue"]
	}
	return template.HTML(fmt.Sprintf(`<div class="p-4 rounded-lg border %s">%s</div>`, cls, text))
}
