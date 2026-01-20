package tunneldebug

// getGatewayDashboardHTML 返回 Gateway 仪表盘 HTML (使用 Tailwind CSS + Alpine.js)
func getGatewayDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel Gateway - Debug Console</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        [x-cloak] { display: none !important; }
        .loading { animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: .5; } }
        @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        .animate-spin { animation: spin 1s linear infinite; }
    </style>
</head>
<body class="bg-gray-900 text-gray-100 min-h-screen">
    <!-- 导航栏 -->
    <nav class="bg-gray-800 border-b border-gray-700 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4">
            <div class="flex items-center justify-between h-14">
                <div class="flex items-center space-x-4">
                    <a href="/debug/" class="text-xl font-bold text-blue-400 hover:text-blue-300">🔧 Debug</a>
                    <span class="text-gray-500">|</span>
                    <span class="text-white font-semibold">🌐 Tunnel Gateway</span>
                </div>
                <div class="flex items-center space-x-4">
                    <a href="/debug/sys" class="text-gray-300 hover:text-white text-sm">系统</a>
                    <a href="/debug/runtime" class="text-gray-300 hover:text-white text-sm">运行时</a>
                    <a href="/debug/pprof/" class="text-gray-300 hover:text-white text-sm">PProf</a>
                </div>
            </div>
        </div>
    </nav>

    <!-- 主内容 -->
    <main class="max-w-7xl mx-auto px-4 py-6" x-data="tunnelDashboard()" x-init="init()">
        <!-- 页面标题 -->
        <div class="mb-6 flex items-center justify-between">
            <div>
                <h1 class="text-2xl font-bold text-white">Tunnel Gateway</h1>
                <p class="text-gray-400 mt-1">服务注册与代理管理</p>
            </div>
            <div class="flex items-center space-x-3">
                <span class="text-gray-500 text-sm">更新于 <span x-text="lastUpdate"></span></span>
                <button @click="refresh()" 
                    class="px-3 py-1.5 rounded text-sm font-medium text-white bg-gray-700 hover:bg-gray-600 transition-colors flex items-center space-x-2"
                    :class="{ 'opacity-50': loading }">
                    <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                    </svg>
                    <span>刷新</span>
                </button>
            </div>
        </div>

        <!-- 统计卡片 -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">注册服务</div>
                <div class="text-2xl font-bold text-white mt-1" x-text="stats.serviceCount">0</div>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">HTTP 端点</div>
                <div class="text-2xl font-bold text-green-400 mt-1" x-text="stats.httpEndpoints">0</div>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">gRPC 端点</div>
                <div class="text-2xl font-bold text-purple-400 mt-1" x-text="stats.grpcEndpoints">0</div>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">Debug 端点</div>
                <div class="text-2xl font-bold text-yellow-400 mt-1" x-text="stats.debugEndpoints">0</div>
            </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <!-- 左侧：Gateway 状态 -->
            <div class="lg:col-span-1">
                <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
                    <div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
                        <h3 class="font-semibold text-white">Gateway 状态</h3>
                        <span class="px-2 py-1 rounded text-xs font-medium"
                            :class="gatewayStatus === 'running' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'">
                            <span class="inline-block w-2 h-2 rounded-full mr-1" 
                                :class="gatewayStatus === 'running' ? 'bg-green-400' : 'bg-red-400'"></span>
                            <span x-text="gatewayStatus === 'running' ? '运行中' : '已停止'"></span>
                        </span>
                    </div>
                    <div class="p-4 space-y-3">
                        <div class="flex items-center justify-between py-2 px-3 bg-gray-700/50 rounded">
                            <div class="flex items-center space-x-2">
                                <span class="text-blue-400">🔗</span>
                                <span class="text-sm text-gray-300">Tunnel</span>
                            </div>
                            <code class="text-sm text-gray-400">:7007</code>
                        </div>
                        <div class="flex items-center justify-between py-2 px-3 bg-gray-700/50 rounded">
                            <div class="flex items-center space-x-2">
                                <span class="text-green-400">🌐</span>
                                <span class="text-sm text-gray-300">HTTP Proxy</span>
                            </div>
                            <code class="text-sm text-gray-400">:8888</code>
                        </div>
                        <div class="flex items-center justify-between py-2 px-3 bg-gray-700/50 rounded">
                            <div class="flex items-center space-x-2">
                                <span class="text-yellow-400">🔧</span>
                                <span class="text-sm text-gray-300">Debug Proxy</span>
                            </div>
                            <code class="text-sm text-gray-400">:6066</code>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 右侧：服务列表 -->
            <div class="lg:col-span-3">
                <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
                    <div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
                        <h3 class="font-semibold text-white">已注册服务</h3>
                        <span class="px-2 py-1 rounded text-xs font-medium bg-blue-500/20 text-blue-400" x-text="services.length + ' 个服务'"></span>
                    </div>
                    <div class="overflow-x-auto">
                        <table class="w-full text-sm">
                            <thead class="bg-gray-700/50">
                                <tr>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium">服务</th>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium">版本</th>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium">状态</th>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium">端点</th>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium">注册时间</th>
                                    <th class="px-4 py-2 text-left text-gray-300 font-medium"></th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-700">
                                <template x-if="services.length === 0">
                                    <tr>
                                        <td colspan="6" class="px-4 py-8 text-center text-gray-500">
                                            <div class="text-4xl mb-2">📭</div>
                                            <div>暂无服务注册</div>
                                        </td>
                                    </tr>
                                </template>
                                <template x-for="svc in services" :key="svc.id">
                                    <tr class="hover:bg-gray-700/30 cursor-pointer" @click="showDetail(svc)">
                                        <td class="px-4 py-3">
                                            <div class="font-medium text-white" x-text="svc.name || '-'"></div>
                                            <div class="text-xs text-gray-500 font-mono" x-text="svc.id ? svc.id.substring(0, 12) : '-'"></div>
                                        </td>
                                        <td class="px-4 py-3">
                                            <code class="text-gray-400 text-xs" x-text="svc.version || '-'"></code>
                                        </td>
                                        <td class="px-4 py-3">
                                            <span class="px-2 py-1 rounded text-xs font-medium"
                                                :class="{
                                                    'bg-green-500/20 text-green-400': svc.status === 'online',
                                                    'bg-red-500/20 text-red-400': svc.status === 'offline',
                                                    'bg-yellow-500/20 text-yellow-400': svc.status === 'unhealthy'
                                                }">
                                                <span class="inline-block w-1.5 h-1.5 rounded-full mr-1"
                                                    :class="{
                                                        'bg-green-400': svc.status === 'online',
                                                        'bg-red-400': svc.status === 'offline',
                                                        'bg-yellow-400': svc.status === 'unhealthy'
                                                    }"></span>
                                                <span x-text="getStatusText(svc.status)"></span>
                                            </span>
                                        </td>
                                        <td class="px-4 py-3">
                                            <div class="flex flex-wrap gap-1">
                                                <template x-for="ep in (svc.endpoints || [])" :key="ep.type + ep.port">
                                                    <span class="px-1.5 py-0.5 rounded text-xs font-medium"
                                                        :class="{
                                                            'bg-green-500/20 text-green-400': ep.type === 'http',
                                                            'bg-purple-500/20 text-purple-400': ep.type === 'grpc',
                                                            'bg-yellow-500/20 text-yellow-400': ep.type === 'debug'
                                                        }"
                                                        x-text="ep.type.toUpperCase()"></span>
                                                </template>
                                            </div>
                                        </td>
                                        <td class="px-4 py-3 text-gray-500 text-xs" x-text="formatTime(svc.register_time)"></td>
                                        <td class="px-4 py-3">
                                            <button class="text-gray-400 hover:text-blue-400" @click.stop="showDetail(svc)">
                                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                                                </svg>
                                            </button>
                                        </td>
                                    </tr>
                                </template>
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- 快速访问 -->
                <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden mt-4">
                    <div class="px-4 py-3 border-b border-gray-700">
                        <h3 class="font-semibold text-white">快速访问</h3>
                    </div>
                    <div class="p-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div class="bg-gray-700/50 rounded p-3">
                            <div class="text-sm text-gray-300 mb-2">HTTP 服务访问</div>
                            <code class="text-xs text-green-400 break-all">curl http://localhost:8888/{服务名}/api/...</code>
                        </div>
                        <div class="bg-gray-700/50 rounded p-3">
                            <div class="text-sm text-gray-300 mb-2">Debug 接口代理</div>
                            <code class="text-xs text-yellow-400 break-all">curl http://localhost:6066/{服务名}/debug/pprof/</code>
                        </div>
                        <div class="bg-gray-700/50 rounded p-3">
                            <div class="text-sm text-gray-300 mb-2">管理界面 API</div>
                            <code class="text-xs text-blue-400 break-all">curl http://localhost:6067/debug/tunnel/api/services</code>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 服务详情模态框 -->
        <div x-show="detailModal" x-cloak
            class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
            @click.self="detailModal = false" @keydown.escape.window="detailModal = false">
            <div class="bg-gray-800 border border-gray-700 rounded-xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col"
                @click.stop>
                <div class="px-6 py-4 border-b border-gray-700 flex items-center justify-between">
                    <h3 class="text-lg font-semibold text-white flex items-center space-x-2">
                        <span>📦</span>
                        <span x-text="selectedService?.name || '服务详情'"></span>
                    </h3>
                    <button @click="detailModal = false" class="text-gray-400 hover:text-white">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                        </svg>
                    </button>
                </div>
                <div class="p-6 overflow-y-auto space-y-6" x-show="selectedService">
                    <!-- 基本信息 -->
                    <div>
                        <h4 class="text-sm font-medium text-gray-400 mb-3 flex items-center space-x-2">
                            <span>ℹ️</span><span>基本信息</span>
                        </h4>
                        <div class="grid grid-cols-2 gap-3">
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">服务名称</div>
                                <div class="text-white mt-1" x-text="selectedService?.name || '-'"></div>
                            </div>
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">版本</div>
                                <code class="text-gray-300 text-sm" x-text="selectedService?.version || '-'"></code>
                            </div>
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">服务 ID</div>
                                <code class="text-gray-400 text-xs break-all" x-text="selectedService?.id || '-'"></code>
                            </div>
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">状态</div>
                                <span class="px-2 py-1 rounded text-xs font-medium mt-1 inline-block"
                                    :class="{
                                        'bg-green-500/20 text-green-400': selectedService?.status === 'online',
                                        'bg-red-500/20 text-red-400': selectedService?.status === 'offline',
                                        'bg-yellow-500/20 text-yellow-400': selectedService?.status === 'unhealthy'
                                    }"
                                    x-text="getStatusText(selectedService?.status)"></span>
                            </div>
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">注册时间</div>
                                <div class="text-gray-300 text-sm mt-1" x-text="formatDateTime(selectedService?.register_time)"></div>
                            </div>
                            <div class="bg-gray-700/50 rounded p-3">
                                <div class="text-xs text-gray-500">最后心跳</div>
                                <div class="text-gray-300 text-sm mt-1" x-text="formatDateTime(selectedService?.last_heartbeat)"></div>
                            </div>
                        </div>
                    </div>

                    <!-- 端点列表 -->
                    <div>
                        <h4 class="text-sm font-medium text-gray-400 mb-3 flex items-center space-x-2">
                            <span>🔌</span>
                            <span>端点列表 (<span x-text="(selectedService?.endpoints || []).length"></span>)</span>
                        </h4>
                        <div class="space-y-2">
                            <template x-for="ep in (selectedService?.endpoints || [])" :key="ep.type + ep.port">
                                <div class="flex items-center justify-between bg-gray-700/50 rounded p-3">
                                    <div class="flex items-center space-x-3">
                                        <span class="w-8 h-8 rounded flex items-center justify-center text-sm"
                                            :class="{
                                                'bg-green-500/20 text-green-400': ep.type === 'http',
                                                'bg-purple-500/20 text-purple-400': ep.type === 'grpc',
                                                'bg-yellow-500/20 text-yellow-400': ep.type === 'debug'
                                            }"
                                            x-text="ep.type === 'http' ? '🌐' : ep.type === 'grpc' ? '⚡' : '🔧'"></span>
                                        <div>
                                            <div class="text-xs font-medium text-gray-300 uppercase" x-text="ep.type"></div>
                                            <code class="text-xs text-gray-500" x-text="(ep.address || 'localhost:' + ep.port) + (ep.path ? ' → ' + ep.path : '')"></code>
                                        </div>
                                    </div>
                                    <button class="text-gray-400 hover:text-blue-400 p-1" @click="copyUrl(selectedService?.name, ep)" title="复制访问地址">
                                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                                        </svg>
                                    </button>
                                </div>
                            </template>
                            <template x-if="!selectedService?.endpoints || selectedService.endpoints.length === 0">
                                <div class="text-center text-gray-500 py-4">暂无端点</div>
                            </template>
                        </div>
                    </div>

                    <!-- 元数据 -->
                    <div x-show="selectedService?.metadata && Object.keys(selectedService.metadata).length > 0">
                        <h4 class="text-sm font-medium text-gray-400 mb-3 flex items-center space-x-2">
                            <span>🏷️</span><span>元数据</span>
                        </h4>
                        <div class="bg-gray-700/50 rounded overflow-hidden">
                            <table class="w-full text-sm">
                                <tbody class="divide-y divide-gray-600">
                                    <template x-for="(v, k) in (selectedService?.metadata || {})" :key="k">
                                        <tr>
                                            <td class="px-3 py-2 text-gray-500 w-1/3" x-text="k"></td>
                                            <td class="px-3 py-2 text-gray-300" x-text="v"></td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- 访问示例 -->
                    <div>
                        <h4 class="text-sm font-medium text-gray-400 mb-3 flex items-center space-x-2">
                            <span>💻</span><span>访问示例</span>
                        </h4>
                        <div class="bg-gray-900 rounded p-4 font-mono text-sm space-y-2">
                            <div class="text-gray-500"># HTTP 服务访问</div>
                            <div class="text-green-400">curl http://localhost:8888/<span x-text="selectedService?.name"></span>/</div>
                            <div class="text-gray-500 mt-3"># Debug 接口访问</div>
                            <div class="text-yellow-400">curl http://localhost:6066/<span x-text="selectedService?.name"></span>/debug/pprof/</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </main>

    <script>
    function tunnelDashboard() {
        return {
            loading: false,
            lastUpdate: '-',
            gatewayStatus: 'running',
            stats: { serviceCount: 0, httpEndpoints: 0, grpcEndpoints: 0, debugEndpoints: 0 },
            services: [],
            detailModal: false,
            selectedService: null,

            init() {
                this.refresh();
                setInterval(() => this.refresh(), 5000);
            },

            async refresh() {
                this.loading = true;
                try {
                    await Promise.all([this.fetchStatus(), this.fetchServices(), this.fetchStats()]);
                    this.lastUpdate = new Date().toLocaleTimeString('zh-CN');
                } catch (e) {
                    console.error('Refresh failed:', e);
                } finally {
                    setTimeout(() => this.loading = false, 300);
                }
            },

            async fetchStatus() {
                const res = await fetch('/debug/tunnel/api/status');
                const data = await res.json();
                if (data.gateway) {
                    this.gatewayStatus = data.gateway.status || 'running';
                }
            },

            async fetchServices() {
                const res = await fetch('/debug/tunnel/api/services');
                const data = await res.json();
                this.services = data.services || [];
                this.stats.serviceCount = data.total || 0;
            },

            async fetchStats() {
                const res = await fetch('/debug/tunnel/api/stats');
                const data = await res.json();
                if (data.services && data.services.endpoints) {
                    this.stats.httpEndpoints = data.services.endpoints.http || 0;
                    this.stats.grpcEndpoints = data.services.endpoints.grpc || 0;
                    this.stats.debugEndpoints = data.services.endpoints.debug || 0;
                }
            },

            showDetail(svc) {
                this.selectedService = svc;
                this.detailModal = true;
            },

            getStatusText(status) {
                const map = { online: '在线', offline: '离线', unhealthy: '异常' };
                return map[status] || '在线';
            },

            formatTime(iso) {
                if (!iso) return '-';
                try { return new Date(iso).toLocaleTimeString('zh-CN'); }
                catch { return '-'; }
            },

            formatDateTime(iso) {
                if (!iso) return '-';
                try { return new Date(iso).toLocaleString('zh-CN'); }
                catch { return '-'; }
            },

            copyUrl(serviceName, ep) {
                const url = ep.type === 'debug' 
                    ? 'http://localhost:6066/' + serviceName + '/debug/'
                    : 'http://localhost:8888/' + serviceName + '/';
                navigator.clipboard.writeText(url);
            }
        }
    }
    </script>
</body>
</html>`
}

// getAgentDashboardHTML 返回 Agent 仪表盘 HTML (使用 Tailwind CSS + Alpine.js)
func getAgentDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel Agent - Debug Console</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
    <style>
        [x-cloak] { display: none !important; }
        .loading { animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: .5; } }
        @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        .animate-spin { animation: spin 1s linear infinite; }
    </style>
</head>
<body class="bg-gray-900 text-gray-100 min-h-screen">
    <!-- 导航栏 -->
    <nav class="bg-gray-800 border-b border-gray-700 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4">
            <div class="flex items-center justify-between h-14">
                <div class="flex items-center space-x-4">
                    <a href="/debug/" class="text-xl font-bold text-blue-400 hover:text-blue-300">🔧 Debug</a>
                    <span class="text-gray-500">|</span>
                    <span class="text-white font-semibold">🔗 Tunnel Agent</span>
                </div>
                <div class="flex items-center space-x-4">
                    <a href="/debug/sys" class="text-gray-300 hover:text-white text-sm">系统</a>
                    <a href="/debug/runtime" class="text-gray-300 hover:text-white text-sm">运行时</a>
                    <a href="/debug/pprof/" class="text-gray-300 hover:text-white text-sm">PProf</a>
                </div>
            </div>
        </div>
    </nav>

    <!-- 主内容 -->
    <main class="max-w-3xl mx-auto px-4 py-6" x-data="agentDashboard()" x-init="init()">
        <!-- 页面标题 -->
        <div class="mb-6 flex items-center justify-between">
            <div>
                <h1 class="text-2xl font-bold text-white">Tunnel Agent</h1>
                <p class="text-gray-400 mt-1">Gateway 连接状态</p>
            </div>
            <div class="flex items-center space-x-3">
                <span class="text-gray-500 text-sm">更新于 <span x-text="lastUpdate"></span></span>
                <button @click="refresh()" 
                    class="px-3 py-1.5 rounded text-sm font-medium text-white bg-gray-700 hover:bg-gray-600 transition-colors flex items-center space-x-2"
                    :class="{ 'opacity-50': loading }">
                    <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                    </svg>
                    <span>刷新</span>
                </button>
            </div>
        </div>

        <!-- 连接状态卡片 -->
        <div class="bg-gray-800 rounded-lg border border-gray-700 p-8 text-center mb-6">
            <div class="w-20 h-20 rounded-full mx-auto flex items-center justify-center text-3xl mb-4"
                :class="{
                    'bg-green-500/20': status === 'connected',
                    'bg-red-500/20': status === 'disconnected',
                    'bg-yellow-500/20 loading': status === 'connecting' || status === 'reconnecting'
                }">
                <span x-text="status === 'connected' ? '✓' : status === 'disconnected' ? '✗' : '⟳'"></span>
            </div>
            <div class="text-2xl font-bold text-white mb-1" x-text="getStatusText()"></div>
            <div class="text-gray-400" x-text="getStatusDesc()"></div>
        </div>

        <!-- 连接信息 -->
        <div class="grid grid-cols-2 gap-4 mb-6">
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">Gateway 地址</div>
                <code class="text-white mt-1 block" x-text="gatewayAddr || '-'"></code>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">服务名称</div>
                <div class="text-white mt-1" x-text="serviceName || '-'"></div>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">服务版本</div>
                <code class="text-white mt-1 block" x-text="serviceVersion || '-'"></code>
            </div>
            <div class="bg-gray-800 rounded-lg border border-gray-700 p-4">
                <div class="text-gray-400 text-sm">端点数量</div>
                <div class="text-white mt-1" x-text="endpoints.length"></div>
            </div>
        </div>

        <!-- 本地端点 -->
        <div class="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <div class="px-4 py-3 border-b border-gray-700">
                <h3 class="font-semibold text-white">本地端点</h3>
            </div>
            <div class="p-4 space-y-2">
                <template x-for="ep in endpoints" :key="ep.type + ep.port">
                    <div class="flex items-center space-x-3 bg-gray-700/50 rounded p-3">
                        <span class="w-9 h-9 rounded-lg flex items-center justify-center"
                            :class="{
                                'bg-green-500/20 text-green-400': ep.type === 'http',
                                'bg-purple-500/20 text-purple-400': ep.type === 'grpc',
                                'bg-yellow-500/20 text-yellow-400': ep.type === 'debug'
                            }"
                            x-text="ep.type === 'http' ? '🌐' : ep.type === 'grpc' ? '⚡' : '🔧'"></span>
                        <div>
                            <div class="text-sm font-medium text-gray-300 uppercase" x-text="ep.type"></div>
                            <code class="text-xs text-gray-500" x-text="ep.address || 'localhost:' + ep.port"></code>
                        </div>
                    </div>
                </template>
                <template x-if="endpoints.length === 0">
                    <div class="text-center text-gray-500 py-4">暂无端点</div>
                </template>
            </div>
        </div>
    </main>

    <script>
    function agentDashboard() {
        return {
            loading: false,
            lastUpdate: '-',
            status: 'disconnected',
            gatewayAddr: '',
            serviceName: '',
            serviceVersion: '',
            endpoints: [],

            init() {
                this.refresh();
                setInterval(() => this.refresh(), 5000);
            },

            async refresh() {
                this.loading = true;
                try {
                    const res = await fetch('/debug/tunnel/api/status');
                    const data = await res.json();
                    if (data.agent) {
                        this.status = data.agent.status || 'disconnected';
                        this.gatewayAddr = data.agent.gateway_addr || '';
                        this.serviceName = data.agent.service_name || '';
                        this.serviceVersion = data.agent.service_version || '';
                        this.endpoints = data.agent.endpoints || [];
                    }
                    this.lastUpdate = new Date().toLocaleTimeString('zh-CN');
                } catch (e) {
                    console.error('Refresh failed:', e);
                } finally {
                    setTimeout(() => this.loading = false, 300);
                }
            },

            getStatusText() {
                const map = {
                    connected: '已连接',
                    connecting: '连接中...',
                    reconnecting: '重连中...',
                    disconnected: '未连接'
                };
                return map[this.status] || '未连接';
            },

            getStatusDesc() {
                const map = {
                    connected: '与 Gateway 连接正常',
                    connecting: '正在连接到 Gateway',
                    reconnecting: '正在尝试重新连接',
                    disconnected: '与 Gateway 断开连接'
                };
                return map[this.status] || '与 Gateway 断开连接';
            }
        }
    }
    </script>
</body>
</html>`
}

// getEmptyDashboardHTML 返回空状态页面
func getEmptyDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel - 未配置</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-900 text-gray-100 min-h-screen flex items-center justify-center">
    <div class="text-center max-w-md px-4">
        <div class="w-20 h-20 rounded-full bg-gray-700/50 mx-auto flex items-center justify-center text-4xl mb-6">
            🌐
        </div>
        <h1 class="text-2xl font-bold text-white mb-2">Tunnel 未配置</h1>
        <p class="text-gray-400">请先配置 Gateway 或 Agent 后再访问此页面</p>
        <a href="/debug/" class="inline-block mt-6 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-white text-sm transition-colors">
            返回 Debug 首页
        </a>
    </div>
</body>
</html>`
}
