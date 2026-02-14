package vars

import (
	"expvar"
	"fmt"
	"html/template"
	"sort"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/recovery"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
)

func init() {
	defer recovery.Exit()

	debug.Route("/vars", func(r fiber.Router) {
		r.Get("/", handleVarsPage)
		r.Get("/api/list", handleVarsList)
		r.Get("/api/get/:name", handleVarGet)
		r.Get("/:name", handleVarDetail)
	})
}

func handleVarsPage(ctx fiber.Ctx) error {
	var vars []varInfo
	expvar.Do(func(kv expvar.KeyValue) {
		vars = append(vars, varInfo{
			Name:  kv.Key,
			Value: kv.Value.String(),
		})
	})
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Name < vars[j].Name
	})

	content := buildVarsContent(vars)

	html, err := ui.Render(ui.PageData{
		Title:       "Expvar 变量",
		Description: "应用运行时导出的变量",
		Breadcrumb:  []string{"Vars"},
		Content:     template.HTML(content),
		ExtraHead:   template.HTML(buildVarsScript()),
	})
	if err != nil {
		return ctx.Status(500).SendString(err.Error())
	}
	ctx.Set("Content-Type", "text/html; charset=utf-8")
	return ctx.SendString(html)
}

type varInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func handleVarsList(ctx fiber.Ctx) error {
	var vars []varInfo
	expvar.Do(func(kv expvar.KeyValue) {
		vars = append(vars, varInfo{
			Name:  kv.Key,
			Value: kv.Value.String(),
		})
	})
	return ctx.JSON(vars)
}

func handleVarGet(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	v := expvar.Get(name)
	if v == nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "variable not found"})
	}
	ctx.Set("Content-Type", "application/json; charset=utf-8")
	return ctx.SendString(v.String())
}

func handleVarDetail(ctx fiber.Ctx) error {
	name := ctx.Params("name")
	v := expvar.Get(name)
	if v == nil {
		return ctx.Status(404).SendString("variable not found")
	}
	ctx.Set("Content-Type", "application/json; charset=utf-8")
	return ctx.SendString(v.String())
}

func buildVarsScript() string {
	return `
<script>
window.varsViewer = function() {
    return {
        vars: [],
        selectedVar: null,
        selectedValue: null,
        searchQuery: '',
        loading: false,

        init() {
            this.loadVars();
        },

        async loadVars() {
            this.loading = true;
            try {
                const res = await fetch('/debug/vars/api/list');
                this.vars = await res.json();
            } catch (e) {
                console.error('Failed to load vars:', e);
            } finally {
                this.loading = false;
            }
        },

        get filteredVars() {
            if (!this.searchQuery) return this.vars;
            const q = this.searchQuery.toLowerCase();
            return this.vars.filter(v => v.name.toLowerCase().includes(q));
        },

        selectVar(v) {
            this.selectedVar = v.name;
            try {
                this.selectedValue = JSON.parse(v.value);
            } catch (e) {
                this.selectedValue = v.value;
            }
        },

        formatJson(obj) {
            if (typeof obj === 'string') return obj;
            return JSON.stringify(obj, null, 2);
        },

        syntaxHighlight(obj) {
            const json = typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2);
            return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
                let cls = 'text-blue-400';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'text-purple-400';
                        match = match.slice(0, -1) + '<span class="text-gray-500">:</span>';
                    } else {
                        cls = 'text-green-400';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'text-yellow-400';
                } else if (/null/.test(match)) {
                    cls = 'text-gray-500';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });
        },

        getPreview(value) {
            try {
                const obj = JSON.parse(value);
                if (typeof obj === 'object' && obj !== null) {
                    const keys = Object.keys(obj);
                    if (keys.length <= 3) {
                        return JSON.stringify(obj);
                    }
                    return '{' + keys.slice(0, 3).join(', ') + ', ...}';
                }
                return String(obj);
            } catch (e) {
                return value.length > 50 ? value.slice(0, 50) + '...' : value;
            }
        },

        getType(value) {
            try {
                const obj = JSON.parse(value);
                if (Array.isArray(obj)) return 'array';
                if (typeof obj === 'object' && obj !== null) return 'object';
                return typeof obj;
            } catch (e) {
                return 'string';
            }
        },

        getTypeColor(value) {
            const type = this.getType(value);
            const colors = {
                'object': 'bg-blue-500/20 text-blue-400',
                'array': 'bg-purple-500/20 text-purple-400',
                'number': 'bg-cyan-500/20 text-cyan-400',
                'string': 'bg-green-500/20 text-green-400',
                'boolean': 'bg-yellow-500/20 text-yellow-400',
            };
            return colors[type] || 'bg-gray-500/20 text-gray-400';
        }
    };
};
</script>
`
}

func buildVarsContent(vars []varInfo) string {
	return fmt.Sprintf(`
<div x-data="varsViewer()" x-init="init()">
    <!-- 统计信息 -->
    <div class="grid grid-cols-3 gap-4 mb-4">
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">总变量数</div>
            <div class="text-2xl font-bold text-white mt-1">%d</div>
        </div>
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">筛选结果</div>
            <div class="text-2xl font-bold text-blue-400 mt-1" x-text="filteredVars.length"></div>
        </div>
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">刷新</div>
            <button @click="loadVars()" class="mt-1 px-3 py-1 bg-blue-600 hover:bg-blue-700 rounded text-white text-sm">
                🔄 刷新数据
            </button>
        </div>
    </div>

    <!-- 搜索框 -->
    <div class="mb-4">
        <input type="text" x-model="searchQuery" placeholder="🔍 搜索变量名..."
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-white">
    </div>

    <!-- 主内容区：左侧列表，右侧详情 -->
    <div class="grid grid-cols-3 gap-4">
        <!-- 变量列表 -->
        <div class="col-span-1 bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <div class="px-4 py-3 border-b border-gray-700 bg-gray-800/80">
                <h3 class="font-medium text-white text-sm">变量列表</h3>
            </div>
            <div class="overflow-auto" style="max-height: calc(100vh - 380px);">
                <template x-for="v in filteredVars" :key="v.name">
                    <div @click="selectVar(v)" 
                        class="px-4 py-3 border-b border-gray-700/50 cursor-pointer transition-colors"
                        :class="selectedVar === v.name ? 'bg-blue-600/20' : 'hover:bg-gray-700/50'">
                        <div class="flex items-center justify-between">
                            <span class="text-white font-medium text-sm" x-text="v.name"></span>
                            <span class="px-1.5 py-0.5 rounded text-xs" :class="getTypeColor(v.value)" x-text="getType(v.value)"></span>
                        </div>
                        <div class="text-gray-500 text-xs mt-1 truncate font-mono" x-text="getPreview(v.value)"></div>
                    </div>
                </template>
                <div x-show="filteredVars.length === 0" class="px-4 py-8 text-center text-gray-500">
                    暂无匹配的变量
                </div>
            </div>
        </div>

        <!-- 变量详情 -->
        <div class="col-span-2 bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <div class="px-4 py-3 border-b border-gray-700 bg-gray-800/80 flex items-center justify-between">
                <div class="flex items-center space-x-2">
                    <h3 class="font-medium text-white text-sm">变量详情</h3>
                    <span x-show="selectedVar" class="text-blue-400 text-sm font-mono" x-text="selectedVar"></span>
                </div>
                <button x-show="selectedVar" @click="navigator.clipboard.writeText(formatJson(selectedValue))" 
                    class="text-xs text-gray-400 hover:text-white px-2 py-1 rounded hover:bg-gray-700">
                    📋 复制
                </button>
            </div>
            <div class="p-4 overflow-auto" style="max-height: calc(100vh - 380px);">
                <template x-if="selectedValue !== null">
                    <pre class="text-sm font-mono whitespace-pre-wrap" x-html="syntaxHighlight(selectedValue)"></pre>
                </template>
                <template x-if="selectedValue === null">
                    <div class="text-center text-gray-500 py-12">
                        <div class="text-4xl mb-3">📊</div>
                        <div>点击左侧变量查看详情</div>
                    </div>
                </template>
            </div>
        </div>
    </div>
</div>
`, len(vars))
}
