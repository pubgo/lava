package loggerdebug

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/closer"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
	"github.com/pubgo/lava/v2/core/logging"
)

// LogEntry 日志条目
type LogEntry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Caller  string         `json:"caller"`
	Message string         `json:"message"`
	Logger  string         `json:"logger"`
	Fields  map[string]any `json:"-"`
	Raw     string         `json:"raw"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Filter   string `json:"filter"`
	Level    string `json:"level"`
	Logger   string `json:"logger"`
	Keyword  string `json:"keyword"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Tail     bool   `json:"tail"`
	FileName string `json:"fileName"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Logs        []LogEntry `json:"logs"`
	Total       int        `json:"total"`
	HasMore     bool       `json:"hasMore"`
	Files       []FileInfo `json:"files"`
	CurrentFile string     `json:"currentFile"`
}

// FileInfo 日志文件信息
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

func init() {
	debug.Route("/logs", func(router fiber.Router) {
		router.Get("/", handlePage)
		router.Get("/api/logs", handleQuery)
		router.Get("/api/files", handleListFiles)
		router.Get("/api/stats", handleStats)
	})
}

// handlePage 渲染日志查看页面
func handlePage(c fiber.Ctx) error {
	logPath := logging.GetLogFilePath()

	html, err := ui.Render(ui.PageData{
		Title:       "日志查看器",
		Description: "查询和分析 JSON 日志文件",
		Breadcrumb:  []string{"Logs"},
		Content:     template.HTML(buildPageContent(logPath)),
		ExtraHead:   template.HTML(buildPageScript()),
	})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

func buildPageScript() string {
	return `
<script>
window.logViewer = function() {
    return {
        logs: [],
        files: [],
        stats: {},
        loggers: [],
        currentFile: '',
        filter: '',
        level: '',
        logger: '',
        keyword: '',
        limit: 100,
        tail: true,
        loading: false,
        hasMore: false,
        detailLog: null,
        detailView: 'form',
        detailParsed: null,

        init() {
            this.loadFiles();
        },

        async loadFiles() {
            try {
                const res = await fetch('/debug/logs/api/files');
                const data = await res.json();
                this.files = data.files || [];
                if (this.files.length > 0 && !this.currentFile) {
                    this.currentFile = this.files[0].name;
                }
                this.loadLogs();
                this.loadStats();
            } catch (e) {
                console.error('Failed to load files:', e);
            }
        },

        async loadLogs() {
            this.loading = true;
            try {
                const params = new URLSearchParams({
                    filter: this.filter,
                    level: this.level,
                    logger: this.logger,
                    keyword: this.keyword,
                    limit: this.limit,
                    tail: this.tail,
                    fileName: this.currentFile,
                });
                const res = await fetch('/debug/logs/api/logs?' + params);
                const data = await res.json();
                this.logs = data.logs || [];
                this.hasMore = data.hasMore;
            } catch (e) {
                console.error('Failed to load logs:', e);
            } finally {
                this.loading = false;
            }
        },

        async loadStats() {
            try {
                const params = new URLSearchParams({ fileName: this.currentFile });
                const res = await fetch('/debug/logs/api/stats?' + params);
                const data = await res.json();
                this.stats = data;
                this.loggers = data.loggers || [];
            } catch (e) {
                console.error('Failed to load stats:', e);
            }
        },

        reset() {
            this.filter = '';
            this.level = '';
            this.logger = '';
            this.keyword = '';
            this.loadLogs();
        },

        showDetail(log) {
            this.detailLog = log;
            this.detailView = 'form';
            try {
                this.detailParsed = JSON.parse(log.raw);
            } catch (e) {
                this.detailParsed = null;
            }
        },

        levelClass(level) {
            const classes = {
                'trace': 'bg-gray-500/20 text-gray-400',
                'debug': 'bg-blue-500/20 text-blue-400',
                'info': 'bg-green-500/20 text-green-400',
                'warn': 'bg-yellow-500/20 text-yellow-400',
                'error': 'bg-red-500/20 text-red-400',
                'fatal': 'bg-red-700/30 text-red-300',
            };
            return classes[level] || classes['info'];
        },

        formatSize(bytes) {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
            return (bytes / 1024 / 1024).toFixed(1) + ' MB';
        },

        formatJson(obj) {
            return JSON.stringify(obj, null, 2);
        },

        isObject(val) {
            return val !== null && typeof val === 'object';
        },

        getValueClass(val) {
            if (val === null) return 'text-gray-500';
            if (typeof val === 'string') return 'text-green-400';
            if (typeof val === 'number') return 'text-blue-400';
            if (typeof val === 'boolean') return 'text-yellow-400';
            return 'text-gray-300';
        },

        formatValue(val) {
            if (val === null) return 'null';
            if (typeof val === 'string') return '"' + val + '"';
            if (typeof val === 'object') return JSON.stringify(val, null, 2);
            return String(val);
        },

        syntaxHighlight(obj) {
            const json = JSON.stringify(obj, null, 2);
            return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
                let cls = 'text-blue-400'; // number
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'text-purple-400'; // key
                        match = match.slice(0, -1) + '<span class="text-gray-500">:</span>';
                    } else {
                        cls = 'text-green-400'; // string
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'text-yellow-400'; // boolean
                } else if (/null/.test(match)) {
                    cls = 'text-gray-500'; // null
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });
        }
    };
};
</script>
`
}

func buildPageContent(logPath string) string {
	return `
<div x-data="logViewer()" x-init="init()">
    <!-- 顶部统计栏 -->
    <div class="flex items-center justify-between mb-4">
        <div class="flex items-center space-x-6">
            <!-- 文件选择 -->
            <div class="flex items-center space-x-2">
                <span class="text-gray-400 text-sm">📁</span>
                <select x-model="currentFile" @change="loadLogs(); loadStats()" 
                    class="bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-white text-sm min-w-48">
                    <template x-for="f in files" :key="f.name">
                        <option :value="f.name" x-text="f.name + ' (' + formatSize(f.size) + ')'"></option>
                    </template>
                </select>
            </div>
            <!-- 统计数字 -->
            <div class="flex items-center space-x-4 text-sm">
                <span class="text-gray-400">总计: <span class="text-white font-medium" x-text="stats.total || '0'"></span></span>
                <span class="text-gray-400">错误: <span class="text-red-400 font-medium" x-text="stats.errors || '0'"></span></span>
                <span class="text-gray-400">警告: <span class="text-yellow-400 font-medium" x-text="stats.warnings || '0'"></span></span>
            </div>
        </div>
        <!-- 刷新和自动刷新 -->
        <div class="flex items-center space-x-3">
            <label class="flex items-center text-sm text-gray-400 cursor-pointer">
                <input type="checkbox" x-model="tail" @change="loadLogs()" class="mr-1.5 rounded">
                最新优先
            </label>
            <button @click="loadLogs(); loadStats()" class="text-gray-400 hover:text-white transition-colors" title="刷新">
                <svg class="w-4 h-4" :class="{'animate-spin': loading}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                </svg>
            </button>
        </div>
    </div>

    <!-- 过滤器 - 两行布局 -->
    <div class="bg-gray-800/50 rounded-lg border border-gray-700 p-4 mb-4">
        <!-- 第一行: 快捷过滤 -->
        <div class="flex items-center space-x-3 mb-3">
            <select x-model="level" @change="loadLogs()" 
                class="bg-gray-700 border border-gray-600 rounded px-3 py-1.5 text-white text-sm">
                <option value="">全部级别</option>
                <option value="trace">Trace</option>
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warn</option>
                <option value="error">Error</option>
                <option value="fatal">Fatal</option>
            </select>
            <select x-model="logger" @change="loadLogs()" 
                class="bg-gray-700 border border-gray-600 rounded px-3 py-1.5 text-white text-sm min-w-32">
                <option value="">全部 Logger</option>
                <template x-for="l in loggers" :key="l">
                    <option :value="l" x-text="l"></option>
                </template>
            </select>
            <div class="flex-1 relative">
                <input type="text" x-model="keyword" @keyup.enter="loadLogs()"
                    placeholder="🔍 关键字搜索..."
                    class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-1.5 text-white text-sm pl-3">
            </div>
            <select x-model="limit" @change="loadLogs()" 
                class="bg-gray-700 border border-gray-600 rounded px-3 py-1.5 text-white text-sm w-20">
                <option value="50">50条</option>
                <option value="100">100条</option>
                <option value="200">200条</option>
                <option value="500">500条</option>
            </select>
            <button @click="loadLogs()" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 rounded text-white text-sm font-medium transition-colors">
                查询
            </button>
            <button @click="reset()" class="px-3 py-1.5 text-gray-400 hover:text-white text-sm transition-colors">
                重置
            </button>
        </div>
        <!-- 第二行: 高级表达式过滤 -->
        <details class="group">
            <summary class="flex items-center cursor-pointer text-sm text-gray-500 hover:text-gray-300">
                <svg class="w-4 h-4 mr-1 transform group-open:rotate-90 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                </svg>
                高级过滤 (expr 表达式)
            </summary>
            <div class="mt-3 space-y-2">
                <input type="text" x-model="filter" @keyup.enter="loadLogs()"
                    placeholder="例如: level == &quot;error&quot; &amp;&amp; contains(message, &quot;timeout&quot;)"
                    class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-white text-sm font-mono">
                <div class="flex flex-wrap gap-2 text-xs">
                    <span class="text-gray-500">可用变量:</span>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-blue-400">level</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-blue-400">message</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-blue-400">logger</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-blue-400">caller</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-blue-400">fields</code>
                    <span class="text-gray-500 ml-2">函数:</span>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-green-400">contains(s, sub)</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-green-400">startsWith(s, pre)</code>
                    <code class="bg-gray-700 px-1.5 py-0.5 rounded text-green-400">endsWith(s, suf)</code>
                </div>
            </div>
        </details>
    </div>

    <!-- 日志列表 -->
    <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
        <div class="px-4 py-2.5 border-b border-gray-700 flex items-center justify-between bg-gray-800/80">
            <div class="flex items-center space-x-3">
                <h3 class="font-medium text-white text-sm">日志列表</h3>
                <span class="text-xs text-gray-500">显示 <span class="text-gray-300" x-text="logs.length"></span> 条</span>
                <span x-show="hasMore" class="text-xs text-blue-400">有更多记录</span>
            </div>
            <span x-show="loading" class="text-xs text-gray-400 loading">加载中...</span>
        </div>
        <div class="overflow-auto" style="max-height: calc(100vh - 280px);">
            <table class="w-full text-sm">
                <thead class="bg-gray-700/50 sticky top-0 z-10">
                    <tr>
                        <th class="px-3 py-2 text-left text-gray-400 font-medium text-xs w-40">时间</th>
                        <th class="px-3 py-2 text-left text-gray-400 font-medium text-xs w-16">级别</th>
                        <th class="px-3 py-2 text-left text-gray-400 font-medium text-xs w-28">Logger</th>
                        <th class="px-3 py-2 text-left text-gray-400 font-medium text-xs">消息</th>
                        <th class="px-3 py-2 text-left text-gray-400 font-medium text-xs w-44">Caller</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-700/50">
                    <template x-for="(log, i) in logs" :key="i">
                        <tr class="hover:bg-gray-700/30 cursor-pointer transition-colors" @click="showDetail(log)"
                            :class="{'bg-red-900/10 hover:bg-red-900/20': log.level === 'error' || log.level === 'fatal', 'bg-yellow-900/10 hover:bg-yellow-900/20': log.level === 'warn'}">
                            <td class="px-3 py-1.5 text-gray-500 font-mono text-xs" x-text="log.time"></td>
                            <td class="px-3 py-1.5">
                                <span class="px-1.5 py-0.5 rounded text-xs font-medium"
                                    :class="levelClass(log.level)" x-text="log.level"></span>
                            </td>
                            <td class="px-3 py-1.5 text-blue-400/80 text-xs truncate" x-text="log.logger || '-'"></td>
                            <td class="px-3 py-1.5 text-gray-200 truncate max-w-lg" x-text="log.message"></td>
                            <td class="px-3 py-1.5 text-gray-600 font-mono text-xs truncate" x-text="log.caller"></td>
                        </tr>
                    </template>
                    <tr x-show="logs.length === 0 && !loading">
                        <td colspan="5" class="px-4 py-12 text-center text-gray-500">
                            <div class="text-3xl mb-2">📭</div>
                            暂无匹配的日志记录
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>

    <!-- 详情弹窗 -->
    <div x-show="detailLog" x-cloak @click.self="detailLog = null" x-transition:enter="transition ease-out duration-200"
        x-transition:enter-start="opacity-0" x-transition:enter-end="opacity-100"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div class="bg-gray-800 rounded-xl border border-gray-700 w-full max-w-4xl max-h-[85vh] overflow-hidden shadow-2xl"
            x-transition:enter="transition ease-out duration-200" x-transition:enter-start="opacity-0 scale-95" x-transition:enter-end="opacity-100 scale-100">
            <!-- 头部 -->
            <div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between bg-gray-800/90">
                <div class="flex items-center space-x-3">
                    <h3 class="font-medium text-white">日志详情</h3>
                    <span class="px-2 py-0.5 rounded text-xs font-medium" :class="levelClass(detailLog?.level)" x-text="detailLog?.level"></span>
                    <span class="text-xs text-gray-500 font-mono" x-text="detailLog?.time"></span>
                </div>
                <div class="flex items-center space-x-2">
                    <button @click="navigator.clipboard.writeText(detailLog?.raw)" class="text-xs text-gray-400 hover:text-white px-2 py-1 rounded hover:bg-gray-700">
                        📋 复制
                    </button>
                    <button @click="detailLog = null" class="text-gray-400 hover:text-white w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-700 transition-colors">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
            </div>
            <!-- Tab 切换 -->
            <div class="px-4 py-2 border-b border-gray-700 flex items-center space-x-1 bg-gray-800/50">
                <button @click="detailView = 'form'" 
                    class="px-3 py-1.5 text-sm rounded-lg transition-colors"
                    :class="detailView === 'form' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-700'">
                    📋 表单视图
                </button>
                <button @click="detailView = 'json'" 
                    class="px-3 py-1.5 text-sm rounded-lg transition-colors"
                    :class="detailView === 'json' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-700'">
                    { } JSON
                </button>
            </div>
            <!-- 内容区 -->
            <div class="p-4 overflow-auto" style="max-height: calc(85vh - 110px);">
                <!-- 表单视图 -->
                <div x-show="detailView === 'form'" x-cloak>
                    <template x-if="detailParsed">
                        <div class="space-y-1">
                            <template x-for="(value, key) in detailParsed" :key="key">
                                <div class="bg-gray-900/50 rounded-lg overflow-hidden">
                                    <div class="flex">
                                        <div class="w-28 flex-shrink-0 px-3 py-2 bg-gray-700/30 text-gray-400 text-sm font-medium" x-text="key"></div>
                                        <div class="flex-1 px-3 py-2 min-w-0">
                                            <template x-if="!isObject(value)">
                                                <span class="text-sm font-mono break-all" :class="getValueClass(value)" x-text="formatValue(value)"></span>
                                            </template>
                                            <template x-if="isObject(value)">
                                                <pre class="text-sm font-mono text-gray-300 whitespace-pre-wrap break-all" x-text="formatJson(value)"></pre>
                                            </template>
                                        </div>
                                    </div>
                                </div>
                            </template>
                        </div>
                    </template>
                    <template x-if="!detailParsed">
                        <div class="text-center text-gray-500 py-8">无法解析 JSON</div>
                    </template>
                </div>
                <!-- 格式化 JSON 视图 -->
                <div x-show="detailView === 'json'" x-cloak>
                    <template x-if="detailParsed">
                        <div class="bg-gray-900 rounded-lg p-4 text-sm font-mono overflow-auto">
                            <pre x-html="syntaxHighlight(detailParsed)"></pre>
                        </div>
                    </template>
                    <template x-if="!detailParsed">
                        <div class="text-center text-gray-500 py-8">无法解析 JSON</div>
                    </template>
                </div>
            </div>
        </div>
    </div>
</div>
`
}

// handleQuery 处理日志查询
func handleQuery(c fiber.Ctx) error {
	limit := 100
	if v, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = v
	}
	if limit <= 0 {
		limit = 100
	}
	offset := 0
	if v, err := strconv.Atoi(c.Query("offset")); err == nil {
		offset = v
	}
	if offset < 0 {
		offset = 0
	}
	tail := true
	if v := strings.TrimSpace(c.Query("tail")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			tail = b
		}
	}

	req := QueryRequest{
		Filter:   c.Query("filter"),
		Level:    c.Query("level"),
		Logger:   c.Query("logger"),
		Keyword:  c.Query("keyword"),
		Start:    c.Query("start"),
		End:      c.Query("end"),
		Limit:    limit,
		Offset:   offset,
		Tail:     tail,
		FileName: c.Query("fileName"),
	}

	if req.Limit > 1000 {
		req.Limit = 1000
	}

	logPath := logging.GetLogFilePath()
	if logPath == "" {
		return c.JSON(QueryResponse{Logs: []LogEntry{}})
	}

	filePath := logPath
	if req.FileName != "" {
		dir := filepath.Dir(logPath)
		filePath = filepath.Join(dir, req.FileName)
	}

	logs, hasMore, err := queryLogs(filePath, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(QueryResponse{
		Logs:        logs,
		Total:       len(logs),
		HasMore:     hasMore,
		CurrentFile: filepath.Base(filePath),
	})
}

// handleListFiles 列出日志文件
func handleListFiles(c fiber.Ctx) error {
	logPath := logging.GetLogFilePath()
	if logPath == "" {
		return c.JSON(fiber.Map{"files": []FileInfo{}})
	}

	dir := filepath.Dir(logPath)
	baseName := filepath.Base(logPath)
	baseNameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	var files []FileInfo
	entries, err := os.ReadDir(dir)
	if err != nil {
		return c.JSON(fiber.Map{"files": files})
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, baseNameNoExt) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Name:    name,
				Size:    info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime > files[j].ModTime
	})

	return c.JSON(fiber.Map{"files": files})
}

// handleStats 获取日志统计
func handleStats(c fiber.Ctx) error {
	logPath := logging.GetLogFilePath()
	if logPath == "" {
		return c.JSON(fiber.Map{})
	}

	fileName := c.Query("fileName")
	filePath := logPath
	if fileName != "" {
		dir := filepath.Dir(logPath)
		filePath = filepath.Join(dir, fileName)
	}

	stats := countLogStats(filePath)
	return c.JSON(stats)
}

// queryLogs 查询日志
func queryLogs(filePath string, req QueryRequest) ([]LogEntry, bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, false, nil
		}
		return nil, false, err
	}
	defer closer.SafeClose(file)

	var exprProgram *vm.Program
	if req.Filter != "" {
		env := map[string]any{
			"level":   "",
			"message": "",
			"logger":  "",
			"caller":  "",
			"time":    "",
			"fields":  map[string]any{},
		}
		prog, err := expr.Compile(req.Filter, expr.Env(env), expr.AsBool())
		if err != nil {
			return nil, false, fmt.Errorf("invalid filter expression: %w", err)
		}
		exprProgram = prog
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var allLogs []LogEntry
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, ok := parseLine(line)
		if !ok {
			continue
		}

		if req.Level != "" && entry.Level != req.Level {
			continue
		}

		if req.Logger != "" && entry.Logger != req.Logger {
			continue
		}

		if req.Keyword != "" && !strings.Contains(strings.ToLower(entry.Raw), strings.ToLower(req.Keyword)) {
			continue
		}

		if exprProgram != nil {
			env := map[string]any{
				"level":   entry.Level,
				"message": entry.Message,
				"logger":  entry.Logger,
				"caller":  entry.Caller,
				"time":    entry.Time,
				"fields":  entry.Fields,
			}
			result, err := expr.Run(exprProgram, env)
			if err != nil || result != true {
				continue
			}
		}

		allLogs = append(allLogs, entry)
	}

	total := len(allLogs)
	hasMore := false
	var logs []LogEntry

	if req.Tail {
		start := total - req.Limit - req.Offset
		if start < 0 {
			start = 0
		}
		end := total - req.Offset
		if end > total {
			end = total
		}
		if end < 0 {
			end = 0
		}
		if start < end {
			logs = allLogs[start:end]
		}
		hasMore = start > 0
		for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
			logs[i], logs[j] = logs[j], logs[i]
		}
	} else {
		start := req.Offset
		end := req.Offset + req.Limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		logs = allLogs[start:end]
		hasMore = end < total
	}

	return logs, hasMore, scanner.Err()
}

// parseLine 解析单行日志
func parseLine(line string) (LogEntry, bool) {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return LogEntry{}, false
	}

	entry := LogEntry{
		Raw:    line,
		Fields: make(map[string]any),
	}

	if v, ok := data["time"].(string); ok {
		entry.Time = v
	}
	if v, ok := data["level"].(string); ok {
		entry.Level = v
	}
	if v, ok := data["caller"].(string); ok {
		entry.Caller = v
	}
	if v, ok := data["message"].(string); ok {
		entry.Message = v
	}
	if v, ok := data["logger"].(string); ok {
		entry.Logger = v
	}

	for k, v := range data {
		if k != "time" && k != "level" && k != "caller" && k != "message" && k != "logger" {
			entry.Fields[k] = v
		}
	}

	return entry, true
}

// LogStats 日志统计
type LogStats struct {
	Total    int      `json:"total"`
	Trace    int      `json:"trace"`
	Debug    int      `json:"debug"`
	Info     int      `json:"info"`
	Warnings int      `json:"warnings"`
	Errors   int      `json:"errors"`
	Loggers  []string `json:"loggers"`
}

// countLogStats 统计日志
func countLogStats(filePath string) LogStats {
	stats := LogStats{}
	loggerSet := make(map[string]struct{})

	file, err := os.Open(filePath)
	if err != nil {
		return stats
	}
	defer closer.SafeClose(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		stats.Total++

		var data struct {
			Level  string `json:"level"`
			Logger string `json:"logger"`
		}
		if json.Unmarshal([]byte(line), &data) == nil {
			switch data.Level {
			case "trace":
				stats.Trace++
			case "debug":
				stats.Debug++
			case "info":
				stats.Info++
			case "warn", "warning":
				stats.Warnings++
			case "error", "fatal", "panic":
				stats.Errors++
			}
			if data.Logger != "" {
				loggerSet[data.Logger] = struct{}{}
			}
		}
	}

	// 转换为排序后的列表
	for logger := range loggerSet {
		stats.Loggers = append(stats.Loggers, logger)
	}
	sort.Strings(stats.Loggers)

	return stats
}

// ReadLastLines 读取文件最后 n 行
func ReadLastLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer closer.SafeClose(file)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	if size == 0 {
		return []string{}, nil
	}

	bufSize := int64(4096)
	if bufSize > size {
		bufSize = size
	}

	var lines []string
	pos := size
	var partial string

	for pos > 0 && len(lines) < n {
		readSize := bufSize
		if pos < bufSize {
			readSize = pos
		}
		pos -= readSize

		buf := make([]byte, readSize)
		_, err := file.ReadAt(buf, pos)
		if err != nil && err != io.EOF {
			return nil, err
		}

		chunk := string(buf) + partial
		partial = ""

		splitLines := strings.Split(chunk, "\n")
		if pos > 0 {
			partial = splitLines[0]
			splitLines = splitLines[1:]
		}

		for i := len(splitLines) - 1; i >= 0; i-- {
			if splitLines[i] != "" {
				lines = append([]string{splitLines[i]}, lines...)
			}
			if len(lines) >= n {
				break
			}
		}
	}

	if partial != "" && len(lines) < n {
		lines = append([]string{partial}, lines...)
	}

	return lines, nil
}
