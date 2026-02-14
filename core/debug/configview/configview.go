package configview

import (
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/config"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password`),
	regexp.MustCompile(`(?i)secret`),
	regexp.MustCompile(`(?i)token`),
	regexp.MustCompile(`(?i)api[_-]?key`),
	regexp.MustCompile(`(?i)access[_-]?key`),
	regexp.MustCompile(`(?i)private[_-]?key`),
	regexp.MustCompile(`(?i)credential`),
	regexp.MustCompile(`(?i)auth`),
	regexp.MustCompile(`(?i)dsn`),
	regexp.MustCompile(`(?i)connection[_-]?string`),
}

func init() {
	debug.Get("/config", func(ctx fiber.Ctx) error {
		configPath := config.GetConfigPath()

		// JSON 响应
		if ctx.Get("Accept") == "application/json" || ctx.Query("format") == "json" {
			if configPath == "" {
				return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "config path not found",
				})
			}

			data, err := os.ReadFile(configPath)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to read config: " + err.Error(),
				})
			}

			var cfg map[string]any
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to parse config: " + err.Error(),
				})
			}

			maskSensitiveData(cfg, "")

			return ctx.JSON(fiber.Map{
				"config_path": configPath,
				"config":      cfg,
			})
		}

		// HTML 页面
		var cfg map[string]any
		var configError string
		var configYaml string

		if configPath == "" {
			configError = "配置文件路径未设置"
		} else {
			data, err := os.ReadFile(configPath)
			if err != nil {
				configError = "读取配置文件失败: " + err.Error()
			} else {
				configYaml = string(data)
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					configError = "解析配置文件失败: " + err.Error()
				} else {
					maskSensitiveData(cfg, "")
				}
			}
		}

		// 统计卡片
		keyCount := countKeys(cfg)
		statsContent := ui.StatsCard("配置路径", shortenPath(configPath, 30), "")
		statsContent += ui.StatsCard("配置项数", fmt.Sprintf("%d", keyCount), "")
		statsContent += ui.StatsCard("敏感字段", fmt.Sprintf("%d 个模式", len(sensitivePatterns)), "自动脱敏")

		// 错误提示
		errorContent := template.HTML("")
		if configError != "" {
			errorContent = ui.Alert(configError, "red")
		}

		// 配置树形展示
		configContent := template.HTML("")
		if cfg != nil {
			configContent = buildConfigTree(cfg, "")
		}

		// 环境变量统计
		envCount := len(os.Environ())

		// 快速操作
		actionsContent := template.HTML(fmt.Sprintf(`
<div class="flex flex-wrap gap-2">
    <a href="/debug/config/path" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">配置路径</a>
    <a href="/debug/config/env" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-green-600 hover:bg-green-700">环境变量 (%d)</a>
    <button onclick="copyConfig()" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-purple-600 hover:bg-purple-700">复制配置</button>
    <button onclick="toggleView()" id="viewBtn" class="px-3 py-1.5 rounded text-sm font-medium text-white bg-gray-600 hover:bg-gray-700">切换视图</button>
</div>
<script>
let viewMode = 'tree';
function toggleView() {
    viewMode = viewMode === 'tree' ? 'yaml' : 'tree';
    document.getElementById('configTree').classList.toggle('hidden');
    document.getElementById('configYaml').classList.toggle('hidden');
    document.getElementById('viewBtn').textContent = viewMode === 'tree' ? '切换到 YAML' : '切换到树形';
}
function copyConfig() {
    const text = document.getElementById('configYaml').textContent;
    navigator.clipboard.writeText(text).then(() => {
        alert('配置已复制到剪贴板');
    });
}
</script>`, envCount))

		// YAML 视图
		yamlView := template.HTML(fmt.Sprintf(`
<pre id="configYaml" class="hidden bg-gray-900 rounded p-4 text-sm text-gray-300 overflow-auto max-h-96 font-mono">%s</pre>`,
			template.HTMLEscapeString(maskYaml(configYaml))))

		content := template.HTML(fmt.Sprintf(`
<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">%s</div>
%s
%s
<div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
    <div class="px-4 py-3 border-b border-gray-700">
        <h3 class="font-semibold text-white">配置内容</h3>
    </div>
    <div class="p-4">
        <div id="configTree" class="space-y-2">%s</div>
        %s
    </div>
</div>`,
			statsContent,
			errorContent,
			ui.Card("快速操作", actionsContent),
			configContent,
			yamlView))

		html, err := ui.Render(ui.PageData{
			Title:       "配置查看",
			Description: "应用配置文件查看（敏感信息已脱敏）",
			Breadcrumb:  []string{"Config"},
			Content:     content,
		})
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		ctx.Set("Content-Type", "text/html; charset=utf-8")
		return ctx.SendString(html)
	})

	debug.Get("/config/path", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"config_path": config.GetConfigPath(),
		})
	})

	debug.Get("/config/raw", func(ctx fiber.Ctx) error {
		if ctx.Query("confirm") != "yes" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "raw config access requires confirm=yes parameter",
				"warning": "raw config may contain sensitive data",
			})
		}

		configPath := config.GetConfigPath()
		if configPath == "" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "config path not found",
			})
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to read config: " + err.Error(),
			})
		}

		ctx.Set("Content-Type", "text/yaml; charset=utf-8")
		return ctx.Send(data)
	})

	debug.Get("/config/env", func(ctx fiber.Ctx) error {
		envs := make(map[string]string)
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				if isSensitiveKey(key) {
					value = maskString(value)
				}
				envs[key] = value
			}
		}

		return ctx.JSON(fiber.Map{
			"count": len(envs),
			"envs":  envs,
		})
	})
}

func maskSensitiveData(data map[string]any, parentKey string) {
	for key, value := range data {
		fullKey := key
		if parentKey != "" {
			fullKey = parentKey + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			maskSensitiveData(v, fullKey)
		case string:
			if isSensitiveKey(key) || isSensitiveKey(fullKey) {
				data[key] = maskString(v)
			}
		case []any:
			for i, item := range v {
				if m, ok := item.(map[string]any); ok {
					maskSensitiveData(m, fullKey)
					v[i] = m
				}
			}
		}
	}
}

func isSensitiveKey(key string) bool {
	for _, pattern := range sensitivePatterns {
		if pattern.MatchString(key) {
			return true
		}
	}
	return false
}

func maskString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	if len(s) <= 8 {
		return s[:1] + "****" + s[len(s)-1:]
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func countKeys(m map[string]any) int {
	count := 0
	for _, v := range m {
		count++
		if sub, ok := v.(map[string]any); ok {
			count += countKeys(sub)
		}
	}
	return count
}

func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

func buildConfigTree(data map[string]any, indent string) template.HTML {
	result := ""
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	for _, key := range keys {
		value := data[key]
		switch v := value.(type) {
		case map[string]any:
			if len(v) == 0 {
				result += fmt.Sprintf(`<div class="%s"><span class="text-blue-400 font-medium">%s:</span> <span class="text-gray-500">{}</span></div>`, indent, key)
			} else {
				result += fmt.Sprintf(`
<details class="%s group" open>
    <summary class="cursor-pointer list-none flex items-center">
        <svg class="w-4 h-4 mr-1 text-gray-500 transform group-open:rotate-90 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
        </svg>
        <span class="text-blue-400 font-medium">%s</span>
        <span class="text-gray-500 text-xs ml-2">(%d)</span>
    </summary>
    <div class="ml-5 mt-1 pl-3 border-l border-gray-700">%s</div>
</details>`, indent, key, len(v), buildConfigTree(v, ""))
			}
		case []any:
			if len(v) == 0 {
				result += fmt.Sprintf(`<div class="%s"><span class="text-purple-400 font-medium">%s:</span> <span class="text-gray-500">[]</span></div>`, indent, key)
			} else {
				result += fmt.Sprintf(`
<details class="%s group">
    <summary class="cursor-pointer list-none flex items-center">
        <svg class="w-4 h-4 mr-1 text-gray-500 transform group-open:rotate-90 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
        </svg>
        <span class="text-purple-400 font-medium">%s</span>
        <span class="text-gray-500 text-xs ml-2">[%d items]</span>
    </summary>
    <div class="ml-5 mt-1 pl-3 border-l border-gray-700">%s</div>
</details>`, indent, key, len(v), buildArrayTree(v, key))
			}
		default:
			valueStr := fmt.Sprintf("%v", v)
			valueClass := "text-green-400"
			if isSensitiveKey(key) {
				valueClass = "text-yellow-400"
			}
			result += fmt.Sprintf(`<div class="%s"><span class="text-gray-300">%s:</span> <span class="%s">%s</span></div>`, indent, key, valueClass, template.HTMLEscapeString(valueStr))
		}
	}
	return template.HTML(result)
}

func buildArrayTree(arr []any, parentKey string) template.HTML {
	result := ""
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			if len(v) == 0 {
				result += fmt.Sprintf(`<div><span class="text-gray-500">[%d]:</span> <span class="text-gray-500">{}</span></div>`, i)
			} else {
				result += fmt.Sprintf(`
<details class="group">
    <summary class="cursor-pointer list-none flex items-center">
        <svg class="w-4 h-4 mr-1 text-gray-500 transform group-open:rotate-90 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
        </svg>
        <span class="text-gray-400">[%d]</span>
    </summary>
    <div class="ml-5 mt-1 pl-3 border-l border-gray-700">%s</div>
</details>`, i, buildConfigTree(v, ""))
			}
		case []any:
			result += fmt.Sprintf(`<div><span class="text-gray-500">[%d]:</span> <span class="text-gray-400">[%d items]</span></div>`, i, len(v))
		default:
			valueStr := fmt.Sprintf("%v", v)
			valueClass := "text-green-400"
			if isSensitiveKey(parentKey) {
				valueClass = "text-yellow-400"
				valueStr = maskString(valueStr)
			}
			result += fmt.Sprintf(`<div><span class="text-gray-500">[%d]:</span> <span class="%s">%s</span></div>`, i, valueClass, template.HTMLEscapeString(valueStr))
		}
	}
	return template.HTML(result)
}

func maskYaml(yaml string) string {
	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		for _, pattern := range sensitivePatterns {
			if pattern.MatchString(line) {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					lines[i] = parts[0] + ": ****"
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}
