package schedulerdebug

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/debug/ui"
	"github.com/pubgo/lava/v2/core/scheduler"
)

func Init(manager scheduler.JobManager) {
	debug.Route("/scheduler", func(router fiber.Router) {
		// 任务列表页面
		router.Get("/", func(ctx *fiber.Ctx) error {
			return renderPage(ctx, manager)
		})

		// API: 获取所有任务
		router.Get("/api/jobs", func(ctx *fiber.Ctx) error {
			return ctx.JSON(manager.ListJobs())
		})

		// API: 获取指定任务详情
		router.Get("/api/jobs/:name", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			job := manager.GetJob(name)
			if job.IsErr() {
				return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": job.GetErr().Error(),
				})
			}
			return ctx.JSON(job.Unwrap())
		})

		// API: 暂停任务
		router.Post("/api/jobs/:name/pause", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			if err := manager.PauseJob(name); err.IsErr() {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.GetErr().Error(),
				})
			}
			return ctx.JSON(fiber.Map{"status": "paused", "name": name})
		})

		// API: 恢复任务
		router.Post("/api/jobs/:name/resume", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			if err := manager.ResumeJob(name); err.IsErr() {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.GetErr().Error(),
				})
			}
			return ctx.JSON(fiber.Map{"status": "resumed", "name": name})
		})

		// API: 删除任务
		router.Delete("/api/jobs/:name", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			if err := manager.DeleteJob(name); err.IsErr() {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.GetErr().Error(),
				})
			}
			return ctx.JSON(fiber.Map{"status": "deleted", "name": name})
		})

		// API: 重载任务
		router.Post("/api/jobs/:name/reload", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			if err := manager.ReloadJob(name); err.IsErr() {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.GetErr().Error(),
				})
			}
			return ctx.JSON(fiber.Map{"status": "reloaded", "name": name})
		})
	})
}

func renderPage(ctx *fiber.Ctx, manager scheduler.JobManager) error {
	jobs := manager.ListJobs()

	// 统计信息
	var runningCount, stoppedCount int
	for _, job := range jobs {
		if job.Status == scheduler.StatusRunning {
			runningCount++
		} else {
			stoppedCount++
		}
	}

	// 构建初始数据
	jobsJSON, _ := json.Marshal(jobs)

	content := buildContent(len(jobs), runningCount, stoppedCount)

	html, err := ui.Render(ui.PageData{
		Title:       "Scheduler",
		Description: "任务调度管理",
		Breadcrumb:  []string{"Scheduler"},
		Content:     content,
		ExtraHead:   extraHead(string(jobsJSON)),
	})
	if err != nil {
		return ctx.Status(500).SendString(err.Error())
	}

	ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
	return ctx.SendString(html)
}

func extraHead(jobsJSON string) template.HTML {
	return template.HTML(fmt.Sprintf(`
<style>
    .job-row { transition: all 0.2s ease; }
    .job-row:hover { background-color: rgba(55, 65, 81, 0.5); }
    .status-dot { width: 8px; height: 8px; border-radius: 50%%; }
    .status-running { background-color: #10b981; animation: pulse-green 2s infinite; }
    .status-stop { background-color: #6b7280; }
    @keyframes pulse-green {
        0%%, 100%% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4); }
        50%% { box-shadow: 0 0 0 4px rgba(16, 185, 129, 0); }
    }
</style>
<script>
window.schedulerApp = function() {
    return {
        jobs: %s,
        loading: false,
        autoRefresh: false,
        refreshInterval: null,
        searchQuery: '',
        statusFilter: 'all',
        showModal: false,
        selectedJob: null,
        notification: { show: false, message: '', type: 'success' },

        init() {
            this.$watch('autoRefresh', (value) => {
                if (value) {
                    this.refreshInterval = setInterval(() => this.refresh(), 5000);
                } else if (this.refreshInterval) {
                    clearInterval(this.refreshInterval);
                    this.refreshInterval = null;
                }
            });
        },

        get filteredJobs() {
            return this.jobs.filter(job => {
                const matchesSearch = job.Spec.Name.toLowerCase().includes(this.searchQuery.toLowerCase());
                const matchesStatus = this.statusFilter === 'all' || job.Status === this.statusFilter;
                return matchesSearch && matchesStatus;
            });
        },

        get runningCount() {
            return this.jobs.filter(j => j.Status === 'running').length;
        },

        get stoppedCount() {
            return this.jobs.filter(j => j.Status !== 'running').length;
        },

        getJobType(job) {
            if (!job) return '未知';
            if (job.Spec?.Cron) return 'Cron: ' + job.Spec.Cron.Expr;
            if (job.Spec?.Ticker) return 'Every: ' + this.formatDuration(job.Spec.Ticker.Dur);
            if (job.Spec?.Once) return 'Once: ' + this.formatDuration(job.Spec.Once.Delay);
            return '未知';
        },

        formatDuration(ns) {
            if (!ns) return '-';
            const ms = ns / 1000000;
            if (ms < 1000) return ms.toFixed(0) + 'ms';
            const s = ms / 1000;
            if (s < 60) return s.toFixed(1) + 's';
            const m = s / 60;
            if (m < 60) return Math.floor(m) + 'm ' + Math.floor(s %% 60) + 's';
            const h = m / 60;
            return Math.floor(h) + 'h ' + Math.floor(m %% 60) + 'm';
        },

        formatTime(ms) {
            if (!ms || ms === 0) return '-';
            const date = new Date(ms);
            return date.toLocaleString('zh-CN', {
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
        },

        decodeResult(result) {
            if (!result) return '';
            try {
                return atob(result);
            } catch (e) {
                return result;
            }
        },

        async refresh() {
            this.loading = true;
            try {
                const res = await fetch('/debug/scheduler/api/jobs');
                this.jobs = await res.json();
            } catch (e) {
                this.showNotification('刷新失败: ' + e.message, 'error');
            }
            this.loading = false;
        },

        async pauseJob(name) {
            try {
                const res = await fetch('/debug/scheduler/api/jobs/' + encodeURIComponent(name) + '/pause', { method: 'POST' });
                if (res.ok) {
                    this.showNotification('任务 ' + name + ' 已暂停', 'success');
                    this.refresh();
                } else {
                    const data = await res.json();
                    this.showNotification(data.error, 'error');
                }
            } catch (e) {
                this.showNotification('操作失败: ' + e.message, 'error');
            }
        },

        async resumeJob(name) {
            try {
                const res = await fetch('/debug/scheduler/api/jobs/' + encodeURIComponent(name) + '/resume', { method: 'POST' });
                if (res.ok) {
                    this.showNotification('任务 ' + name + ' 已恢复', 'success');
                    this.refresh();
                } else {
                    const data = await res.json();
                    this.showNotification(data.error, 'error');
                }
            } catch (e) {
                this.showNotification('操作失败: ' + e.message, 'error');
            }
        },

        async reloadJob(name) {
            try {
                const res = await fetch('/debug/scheduler/api/jobs/' + encodeURIComponent(name) + '/reload', { method: 'POST' });
                if (res.ok) {
                    this.showNotification('任务 ' + name + ' 已重载', 'success');
                    this.refresh();
                } else {
                    const data = await res.json();
                    this.showNotification(data.error, 'error');
                }
            } catch (e) {
                this.showNotification('操作失败: ' + e.message, 'error');
            }
        },

        async deleteJob(name) {
            if (!confirm('确定要删除任务 ' + name + ' 吗？')) return;
            try {
                const res = await fetch('/debug/scheduler/api/jobs/' + encodeURIComponent(name), { method: 'DELETE' });
                if (res.ok) {
                    this.showNotification('任务 ' + name + ' 已删除', 'success');
                    this.refresh();
                } else {
                    const data = await res.json();
                    this.showNotification(data.error, 'error');
                }
            } catch (e) {
                this.showNotification('操作失败: ' + e.message, 'error');
            }
        },

        showDetail(job) {
            this.selectedJob = job;
            this.showModal = true;
        },

        showError(job) {
            this.selectedJob = { name: job.Spec.Name, error: job.Error };
            this.showModal = true;
        },

        showNotification(message, type = 'success') {
            this.notification = { show: true, message, type };
            setTimeout(() => this.notification.show = false, 3000);
        }
    };
};
</script>`, jobsJSON))
}

func buildContent(total, running, stopped int) template.HTML {
	return template.HTML(fmt.Sprintf(`
<div x-data="schedulerApp()" x-init="init()">
    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">总任务数</div>
            <div class="text-2xl font-bold text-white mt-1" x-text="jobs.length">%d</div>
        </div>
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">运行中</div>
            <div class="text-2xl font-bold text-green-400 mt-1" x-text="runningCount">%d</div>
        </div>
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">已停止</div>
            <div class="text-2xl font-bold text-gray-400 mt-1" x-text="stoppedCount">%d</div>
        </div>
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <div class="text-gray-400 text-sm">自动刷新</div>
            <div class="mt-2">
                <label class="flex items-center cursor-pointer">
                    <input type="checkbox" x-model="autoRefresh" class="sr-only peer">
                    <div class="relative w-11 h-6 bg-gray-600 rounded-full peer peer-checked:bg-green-600 peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all"></div>
                    <span class="ml-2 text-sm text-gray-400" x-text="autoRefresh ? '开启 (5s)' : '关闭'"></span>
                </label>
            </div>
        </div>
    </div>

    <!-- 操作栏 -->
    <div class="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-4">
        <div class="flex items-center justify-between flex-wrap gap-4">
            <div class="flex items-center space-x-4">
                <input type="text" x-model="searchQuery" placeholder="搜索任务名称..."
                    class="bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-sm text-white placeholder-gray-400 focus:outline-none focus:border-blue-500 w-64">
                <select x-model="statusFilter" class="bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-sm text-white focus:outline-none focus:border-blue-500">
                    <option value="all">全部状态</option>
                    <option value="running">运行中</option>
                    <option value="stop">已停止</option>
                </select>
            </div>
            <button @click="refresh()" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-sm font-medium text-white transition-colors flex items-center space-x-2">
                <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                </svg>
                <span>刷新</span>
            </button>
        </div>
    </div>

    <!-- 任务列表 -->
    <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
        <div class="overflow-x-auto">
            <table class="w-full text-sm">
                <thead class="bg-gray-700/50">
                    <tr>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">状态</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">任务名称</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">类型</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">执行次数</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">上次执行</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">下次执行</th>
                        <th class="px-4 py-3 text-left text-gray-300 font-medium">最近结果</th>
                        <th class="px-4 py-3 text-right text-gray-300 font-medium">操作</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-700">
                    <template x-for="job in filteredJobs" :key="job.Spec.Name">
                        <tr class="job-row">
                            <td class="px-4 py-3">
                                <div class="flex items-center space-x-2">
                                    <div class="status-dot" :class="job.Status === 'running' ? 'status-running' : 'status-stop'"></div>
                                    <span class="text-xs font-medium" :class="job.Status === 'running' ? 'text-green-400' : 'text-gray-400'" x-text="job.Status === 'running' ? '运行中' : '已停止'"></span>
                                </div>
                            </td>
                            <td class="px-4 py-3">
                                <span class="text-white font-medium" x-text="job.Spec.Name"></span>
                            </td>
                            <td class="px-4 py-3">
                                <span class="px-2 py-1 rounded text-xs font-medium" 
                                    :class="{
                                        'bg-blue-500/20 text-blue-400': job.Spec.Cron,
                                        'bg-purple-500/20 text-purple-400': job.Spec.Ticker,
                                        'bg-yellow-500/20 text-yellow-400': job.Spec.Once
                                    }"
                                    x-text="getJobType(job)"></span>
                            </td>
                            <td class="px-4 py-3 text-gray-300" x-text="job.Runs"></td>
                            <td class="px-4 py-3 text-gray-300" x-text="formatTime(job.PreExecTime)"></td>
                            <td class="px-4 py-3 text-gray-300" x-text="formatTime(job.ExecTime)"></td>
                            <td class="px-4 py-3">
                                <template x-if="job.Error">
                                    <span class="px-2 py-1 rounded text-xs font-medium bg-red-500/20 text-red-400 cursor-pointer" 
                                        @click="showError(job)" x-text="'错误'"></span>
                                </template>
                                <template x-if="!job.Error && job.Runs > 0">
                                    <span class="px-2 py-1 rounded text-xs font-medium bg-green-500/20 text-green-400">成功</span>
                                </template>
                                <template x-if="!job.Error && job.Runs === 0">
                                    <span class="px-2 py-1 rounded text-xs font-medium bg-gray-500/20 text-gray-400">未执行</span>
                                </template>
                            </td>
                            <td class="px-4 py-3 text-right">
                                <div class="flex items-center justify-end space-x-2">
                                    <button @click="showDetail(job)" class="p-1.5 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors" title="详情">
                                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                                        </svg>
                                    </button>
                                    <template x-if="job.Status === 'running'">
                                        <button @click="pauseJob(job.Spec.Name)" class="p-1.5 text-yellow-400 hover:text-yellow-300 hover:bg-gray-700 rounded transition-colors" title="暂停">
                                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                            </svg>
                                        </button>
                                    </template>
                                    <template x-if="job.Status !== 'running'">
                                        <button @click="resumeJob(job.Spec.Name)" class="p-1.5 text-green-400 hover:text-green-300 hover:bg-gray-700 rounded transition-colors" title="恢复">
                                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                            </svg>
                                        </button>
                                    </template>
                                    <button @click="reloadJob(job.Spec.Name)" class="p-1.5 text-blue-400 hover:text-blue-300 hover:bg-gray-700 rounded transition-colors" title="重载">
                                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                                        </svg>
                                    </button>
                                    <button @click="deleteJob(job.Spec.Name)" class="p-1.5 text-red-400 hover:text-red-300 hover:bg-gray-700 rounded transition-colors" title="删除">
                                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                                        </svg>
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </template>
                    <template x-if="filteredJobs.length === 0">
                        <tr>
                            <td colspan="8" class="px-4 py-8 text-center text-gray-500">
                                <svg class="w-12 h-12 mx-auto mb-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
                                </svg>
                                <p>暂无任务</p>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>
    </div>

    <!-- 详情模态框 -->
    <div x-show="showModal" x-cloak class="fixed inset-0 z-50 overflow-y-auto" @keydown.escape.window="showModal = false">
        <div class="flex items-center justify-center min-h-screen px-4 py-8">
            <div class="fixed inset-0 bg-black/60 backdrop-blur-sm transition-opacity" @click="showModal = false"></div>
            <div class="relative bg-gray-800 rounded-xl border border-gray-700 max-w-3xl w-full max-h-[85vh] overflow-hidden shadow-2xl"
                x-show="showModal"
                x-transition:enter="transition ease-out duration-200"
                x-transition:enter-start="opacity-0 scale-95"
                x-transition:enter-end="opacity-100 scale-100"
                x-transition:leave="transition ease-in duration-150"
                x-transition:leave-start="opacity-100 scale-100"
                x-transition:leave-end="opacity-0 scale-95">
                <!-- 头部 -->
                <div class="px-6 py-4 border-b border-gray-700 flex items-center justify-between bg-gray-800/80">
                    <div class="flex items-center space-x-3">
                        <div class="p-2 rounded-lg bg-blue-500/20">
                            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
                            </svg>
                        </div>
                        <div>
                            <h3 class="text-lg font-semibold text-white" x-text="selectedJob?.Spec?.Name || '任务详情'"></h3>
                            <p class="text-xs text-gray-400" x-text="getJobType(selectedJob)"></p>
                        </div>
                    </div>
                    <button @click="showModal = false" class="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
                
                <!-- 内容 -->
                <div class="p-6 overflow-y-auto max-h-[calc(85vh-80px)]" x-show="selectedJob">
                    <!-- 状态概览 -->
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                        <div class="bg-gray-700/50 rounded-lg p-3 text-center">
                            <div class="text-xs text-gray-400 mb-1">状态</div>
                            <div class="flex items-center justify-center space-x-2">
                                <div class="status-dot" :class="selectedJob?.Status === 'running' ? 'status-running' : 'status-stop'"></div>
                                <span class="text-sm font-medium" :class="selectedJob?.Status === 'running' ? 'text-green-400' : 'text-gray-400'" 
                                    x-text="selectedJob?.Status === 'running' ? '运行中' : '已停止'"></span>
                            </div>
                        </div>
                        <div class="bg-gray-700/50 rounded-lg p-3 text-center">
                            <div class="text-xs text-gray-400 mb-1">执行次数</div>
                            <div class="text-lg font-bold text-white" x-text="selectedJob?.Runs || 0"></div>
                        </div>
                        <div class="bg-gray-700/50 rounded-lg p-3 text-center">
                            <div class="text-xs text-gray-400 mb-1">上次执行</div>
                            <div class="text-sm text-white" x-text="formatTime(selectedJob?.PreExecTime) || '-'"></div>
                        </div>
                        <div class="bg-gray-700/50 rounded-lg p-3 text-center">
                            <div class="text-xs text-gray-400 mb-1">下次执行</div>
                            <div class="text-sm text-white" x-text="formatTime(selectedJob?.ExecTime) || '-'"></div>
                        </div>
                    </div>

                    <!-- 调度配置 -->
                    <div class="bg-gray-700/30 rounded-lg p-4 mb-4">
                        <h4 class="text-sm font-medium text-gray-300 mb-3 flex items-center">
                            <svg class="w-4 h-4 mr-2 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                            </svg>
                            调度配置
                        </h4>
                        <div class="grid grid-cols-2 gap-3 text-sm">
                            <template x-if="selectedJob?.Spec?.Cron">
                                <div class="col-span-2 flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">Cron 表达式</span>
                                    <code class="px-2 py-1 bg-gray-800 rounded text-blue-400" x-text="selectedJob.Spec.Cron.Expr"></code>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Ticker">
                                <div class="col-span-2 flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">执行间隔</span>
                                    <span class="text-white" x-text="formatDuration(selectedJob.Spec.Ticker.Dur)"></span>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Once">
                                <div class="col-span-2 flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">延迟执行</span>
                                    <span class="text-white" x-text="formatDuration(selectedJob.Spec.Once.Delay)"></span>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Config">
                                <div class="flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">超时时间</span>
                                    <span class="text-white" x-text="formatDuration(selectedJob.Spec.Config.Timeout)"></span>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Config">
                                <div class="flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">最大重试</span>
                                    <span class="text-white" x-text="selectedJob.Spec.Config.MaxRetries + ' 次'"></span>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Config">
                                <div class="flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">重试间隔</span>
                                    <span class="text-white" x-text="formatDuration(selectedJob.Spec.Config.RetryInterval)"></span>
                                </div>
                            </template>
                            <template x-if="selectedJob?.Spec?.Config">
                                <div class="flex items-center justify-between py-2 border-b border-gray-600">
                                    <span class="text-gray-400">时区</span>
                                    <span class="text-white" x-text="selectedJob.Spec.Config.Location || 'UTC'"></span>
                                </div>
                            </template>
                        </div>
                    </div>

                    <!-- 执行结果 -->
                    <div class="bg-gray-700/30 rounded-lg p-4 mb-4">
                        <h4 class="text-sm font-medium text-gray-300 mb-3 flex items-center">
                            <svg class="w-4 h-4 mr-2" :class="selectedJob?.Error ? 'text-red-400' : 'text-green-400'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                            </svg>
                            最近执行结果
                        </h4>
                        <template x-if="selectedJob?.Error">
                            <div class="bg-red-500/10 border border-red-500/30 rounded-lg p-3">
                                <div class="flex items-start space-x-2">
                                    <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                    </svg>
                                    <pre class="text-sm text-red-300 whitespace-pre-wrap break-all" x-text="selectedJob.Error"></pre>
                                </div>
                            </div>
                        </template>
                        <template x-if="!selectedJob?.Error && selectedJob?.Result">
                            <div class="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
                                <pre class="text-sm text-green-300 whitespace-pre-wrap break-all" x-text="decodeResult(selectedJob.Result)"></pre>
                            </div>
                        </template>
                        <template x-if="!selectedJob?.Error && !selectedJob?.Result && selectedJob?.Runs > 0">
                            <div class="text-gray-400 text-sm">执行成功，无返回数据</div>
                        </template>
                        <template x-if="selectedJob?.Runs === 0">
                            <div class="text-gray-500 text-sm">尚未执行</div>
                        </template>
                    </div>

                    <!-- 原始数据 -->
                    <details class="group">
                        <summary class="flex items-center justify-between cursor-pointer text-sm text-gray-400 hover:text-gray-300 py-2">
                            <span class="flex items-center">
                                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path>
                                </svg>
                                查看原始 JSON
                            </span>
                            <svg class="w-4 h-4 transform group-open:rotate-180 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                            </svg>
                        </summary>
                        <pre class="mt-2 bg-gray-900 rounded-lg p-4 text-xs text-gray-400 overflow-auto max-h-64" x-text="JSON.stringify(selectedJob, null, 2)"></pre>
                    </details>
                </div>

                <!-- 底部操作 -->
                <div class="px-6 py-4 border-t border-gray-700 bg-gray-800/80 flex items-center justify-between">
                    <div class="flex items-center space-x-2">
                        <template x-if="selectedJob?.Status === 'running'">
                            <button @click="pauseJob(selectedJob.Spec.Name); showModal = false" 
                                class="px-3 py-1.5 bg-yellow-600 hover:bg-yellow-700 rounded-lg text-sm font-medium text-white transition-colors flex items-center space-x-1">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                </svg>
                                <span>暂停</span>
                            </button>
                        </template>
                        <template x-if="selectedJob?.Status !== 'running'">
                            <button @click="resumeJob(selectedJob.Spec.Name); showModal = false"
                                class="px-3 py-1.5 bg-green-600 hover:bg-green-700 rounded-lg text-sm font-medium text-white transition-colors flex items-center space-x-1">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                </svg>
                                <span>恢复</span>
                            </button>
                        </template>
                        <button @click="reloadJob(selectedJob.Spec.Name); showModal = false"
                            class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 rounded-lg text-sm font-medium text-white transition-colors flex items-center space-x-1">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                            </svg>
                            <span>重载</span>
                        </button>
                    </div>
                    <button @click="showModal = false" class="px-4 py-1.5 bg-gray-600 hover:bg-gray-700 rounded-lg text-sm font-medium text-white transition-colors">
                        关闭
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- 通知 -->
    <div x-show="notification.show" x-cloak
        x-transition:enter="transition ease-out duration-300"
        x-transition:enter-start="opacity-0 transform translate-y-2"
        x-transition:enter-end="opacity-100 transform translate-y-0"
        x-transition:leave="transition ease-in duration-200"
        x-transition:leave-start="opacity-100 transform translate-y-0"
        x-transition:leave-end="opacity-0 transform translate-y-2"
        class="fixed bottom-4 right-4 z-50">
        <div class="px-4 py-3 rounded-lg shadow-lg flex items-center space-x-3"
            :class="{
                'bg-green-600': notification.type === 'success',
                'bg-red-600': notification.type === 'error',
                'bg-blue-600': notification.type === 'info'
            }">
            <span class="text-white text-sm" x-text="notification.message"></span>
        </div>
    </div>
</div>`, total, running, stopped))
}
