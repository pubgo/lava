package tunneldebug

// getGatewayDashboardHTML 返回 Gateway 仪表盘 HTML
func getGatewayDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel Gateway 管理界面</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        :root {
            --primary-color: #6366f1;
            --success-color: #22c55e;
            --warning-color: #f59e0b;
            --danger-color: #ef4444;
            --bg-dark: #1e1e2e;
            --bg-card: #282a36;
            --text-primary: #f8f8f2;
            --text-secondary: #6272a4;
        }
        
        body {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            color: var(--text-primary);
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
        }
        
        .navbar {
            background: rgba(30, 30, 46, 0.95) !important;
            backdrop-filter: blur(10px);
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .navbar-brand { font-weight: 700; font-size: 1.5rem; }
        
        .card {
            background: var(--bg-card);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
        }
        
        .card-header {
            background: transparent;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            font-weight: 600;
        }
        
        .stat-card {
            background: linear-gradient(135deg, var(--bg-card) 0%, rgba(99, 102, 241, 0.1) 100%);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        
        .stat-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px -5px rgba(99, 102, 241, 0.3);
        }
        
        .stat-value {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #6366f1, #8b5cf6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .stat-label {
            color: var(--text-secondary);
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.875rem;
            font-weight: 500;
        }
        
        .status-running { background: rgba(34, 197, 94, 0.2); color: var(--success-color); }
        .status-stopped { background: rgba(239, 68, 68, 0.2); color: var(--danger-color); }
        
        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        
        .status-running .status-dot { background: var(--success-color); }
        .status-stopped .status-dot { background: var(--danger-color); }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        .table { color: var(--text-primary); }
        
        .table thead th {
            background: rgba(99, 102, 241, 0.1);
            border-bottom: 2px solid rgba(99, 102, 241, 0.3);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
        }
        
        .table tbody tr {
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
            transition: background 0.2s;
        }
        
        .table tbody tr:hover { background: rgba(99, 102, 241, 0.1); }
        
        .btn-primary {
            background: linear-gradient(135deg, #6366f1, #8b5cf6);
            border: none;
        }
        
        .btn-primary:hover { background: linear-gradient(135deg, #5558e3, #7c4fe8); }
        
        .endpoint-badge {
            display: inline-block;
            padding: 4px 8px;
            margin: 2px;
            border-radius: 6px;
            font-size: 0.75rem;
            font-weight: 500;
        }
        
        .endpoint-http { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
        .endpoint-grpc { background: rgba(99, 102, 241, 0.2); color: #6366f1; }
        .endpoint-debug { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
        
        .refresh-btn { transition: transform 0.3s; }
        .refresh-btn.spinning { animation: spin 1s linear infinite; }
        
        @keyframes spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }
        
        .modal-content {
            background: var(--bg-card);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .modal-header { border-bottom: 1px solid rgba(255, 255, 255, 0.1); }
        .modal-footer { border-top: 1px solid rgba(255, 255, 255, 0.1); }
        
        pre {
            background: var(--bg-dark);
            border-radius: 8px;
            padding: 16px;
            color: #f8f8f2;
            font-size: 0.875rem;
        }
        
        .arch-diagram {
            background: var(--bg-dark);
            border-radius: 12px;
            padding: 20px;
            font-family: 'Fira Code', monospace;
            font-size: 0.75rem;
            line-height: 1.4;
            overflow-x: auto;
        }
        
        .text-muted { color: var(--text-secondary) !important; }
        .last-update { font-size: 0.75rem; color: var(--text-secondary); }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark sticky-top">
        <div class="container-fluid">
            <a class="navbar-brand" href="#">
                <i class="bi bi-hdd-network me-2"></i>Tunnel Gateway
            </a>
            <div class="d-flex align-items-center">
                <span class="last-update me-3">最后更新: <span id="lastUpdate">-</span></span>
                <button class="btn btn-outline-light btn-sm" onclick="refreshData()">
                    <i class="bi bi-arrow-clockwise refresh-btn" id="refreshIcon"></i> 刷新
                </button>
            </div>
        </div>
    </nav>

    <div class="container-fluid py-4">
        <div class="row g-4 mb-4">
            <div class="col-md-3">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="serviceCount">0</div>
                        <div class="stat-label">注册服务</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="httpEndpoints">0</div>
                        <div class="stat-label">HTTP 端点</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="grpcEndpoints">0</div>
                        <div class="stat-label">gRPC 端点</div>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="debugEndpoints">0</div>
                        <div class="stat-label">Debug 端点</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row g-4">
            <div class="col-lg-4">
                <div class="card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <span><i class="bi bi-server me-2"></i>Gateway 状态</span>
                        <span class="status-badge status-running" id="gatewayStatus">
                            <span class="status-dot"></span>
                            <span id="gatewayStatusText">检测中...</span>
                        </span>
                    </div>
                    <div class="card-body">
                        <div class="mb-3">
                            <label class="text-muted small">监听端口</label>
                            <div class="d-flex flex-wrap gap-2 mt-1">
                                <span class="badge bg-primary">:7007 Tunnel</span>
                                <span class="badge bg-success">:8888 HTTP</span>
                                <span class="badge bg-warning text-dark">:6066 Debug</span>
                            </div>
                        </div>
                        <div class="arch-diagram">
                            <pre class="mb-0" style="background: transparent; padding: 0;">
      外部请求
          │
          ▼
┌─────────────────────┐
│   Gateway :7007     │
│ ┌─────┐┌─────┐┌───┐ │
│ │HTTP ││gRPC ││DBG│ │
│ │8888 ││9999 ││6066│ │
│ └──┬──┘└──┬──┘└─┬─┘ │
└────┼──────┼─────┼───┘
     └──────┼─────┘
            ▼
      Agent 连接</pre>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-8">
                <div class="card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <span><i class="bi bi-boxes me-2"></i>已注册服务</span>
                        <span class="badge bg-primary" id="serviceBadge">0 个服务</span>
                    </div>
                    <div class="card-body p-0">
                        <div class="table-responsive">
                            <table class="table table-hover mb-0">
                                <thead>
                                    <tr>
                                        <th>服务名称</th>
                                        <th>版本</th>
                                        <th>状态</th>
                                        <th>端点</th>
                                        <th>操作</th>
                                    </tr>
                                </thead>
                                <tbody id="servicesTable">
                                    <tr>
                                        <td colspan="5" class="text-center text-muted py-4">
                                            <i class="bi bi-hourglass-split me-2"></i>加载中...
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row g-4 mt-2">
            <div class="col-12">
                <div class="card">
                    <div class="card-header"><i class="bi bi-link-45deg me-2"></i>访问指南</div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">HTTP 服务访问</h6>
                                <pre class="mb-0">curl http://localhost:8888/{服务名}/path</pre>
                            </div>
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">Debug 接口访问</h6>
                                <pre class="mb-0">curl http://localhost:6066/{服务名}/debug/pprof/</pre>
                            </div>
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">服务列表 API</h6>
                                <pre class="mb-0">curl http://localhost:8888/</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <div class="modal fade" id="serviceModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title"><i class="bi bi-box me-2"></i>服务详情</h5>
                    <button type="button" class="btn-close btn-close-white" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <div id="serviceDetail">
                        <div class="text-center py-4"><div class="spinner-border text-primary"></div></div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let refreshInterval;

        document.addEventListener('DOMContentLoaded', () => {
            refreshData();
            refreshInterval = setInterval(refreshData, 5000);
        });

        async function refreshData() {
            const icon = document.getElementById('refreshIcon');
            icon.classList.add('spinning');
            try {
                await Promise.all([fetchStatus(), fetchServices(), fetchStats()]);
                document.getElementById('lastUpdate').textContent = new Date().toLocaleTimeString();
            } catch (err) {
                console.error('Failed to refresh:', err);
            } finally {
                setTimeout(() => icon.classList.remove('spinning'), 500);
            }
        }

        async function fetchStatus() {
            try {
                const res = await fetch('/debug/tunnel/api/status');
                const data = await res.json();
                const statusEl = document.getElementById('gatewayStatus');
                const textEl = document.getElementById('gatewayStatusText');
                if (data.gateway) {
                    textEl.textContent = data.gateway.status;
                    statusEl.className = 'status-badge status-' + data.gateway.status;
                } else {
                    textEl.textContent = '未配置';
                    statusEl.className = 'status-badge status-stopped';
                }
            } catch (err) { console.error(err); }
        }

        async function fetchServices() {
            try {
                const res = await fetch('/debug/tunnel/api/services');
                const data = await res.json();
                document.getElementById('serviceCount').textContent = data.total || 0;
                document.getElementById('serviceBadge').textContent = (data.total || 0) + ' 个服务';
                const tbody = document.getElementById('servicesTable');
                if (!data.services || data.services.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted py-4"><i class="bi bi-inbox me-2"></i>暂无服务注册</td></tr>';
                    return;
                }
                tbody.innerHTML = data.services.map(svc => {
                    const endpoints = (svc.endpoints || []).map(ep => 
                        '<span class="endpoint-badge endpoint-' + ep.type + '">' + ep.type + '</span>'
                    ).join('');
                    return '<tr>' +
                        '<td><strong>' + svc.name + '</strong></td>' +
                        '<td><code>' + (svc.version || '-') + '</code></td>' +
                        '<td><span class="status-badge status-running"><span class="status-dot"></span>' + (svc.status || 'active') + '</span></td>' +
                        '<td>' + (endpoints || '-') + '</td>' +
                        '<td><button class="btn btn-sm btn-outline-primary" onclick="viewService(\'' + svc.name + '\')"><i class="bi bi-eye"></i></button></td>' +
                        '</tr>';
                }).join('');
            } catch (err) { console.error(err); }
        }

        async function fetchStats() {
            try {
                const res = await fetch('/debug/tunnel/api/stats');
                const data = await res.json();
                if (data.services && data.services.endpoints) {
                    document.getElementById('httpEndpoints').textContent = data.services.endpoints.http || 0;
                    document.getElementById('grpcEndpoints').textContent = data.services.endpoints.grpc || 0;
                    document.getElementById('debugEndpoints').textContent = data.services.endpoints.debug || 0;
                }
            } catch (err) { console.error(err); }
        }

        async function viewService(name) {
            const modal = new bootstrap.Modal(document.getElementById('serviceModal'));
            const detail = document.getElementById('serviceDetail');
            detail.innerHTML = '<div class="text-center py-4"><div class="spinner-border text-primary"></div></div>';
            modal.show();
            try {
                const res = await fetch('/debug/tunnel/api/services/' + name);
                const svc = await res.json();
                if (svc.error) {
                    detail.innerHTML = '<div class="alert alert-danger">' + svc.error + '</div>';
                    return;
                }
                const endpoints = (svc.endpoints || []).map(ep =>
                    '<tr><td><span class="endpoint-badge endpoint-' + ep.type + '">' + ep.type + '</span></td><td><code>' + ep.address + '</code></td><td>' + (ep.path || '/') + '</td></tr>'
                ).join('');
                const metadata = svc.metadata ? Object.entries(svc.metadata).map(([k, v]) => '<tr><td>' + k + '</td><td>' + v + '</td></tr>').join('') : '<tr><td colspan="2" class="text-muted">无</td></tr>';
                detail.innerHTML = 
                    '<div class="row"><div class="col-md-6"><h6 class="text-muted">基本信息</h6><table class="table table-sm">' +
                    '<tr><td>服务名</td><td><strong>' + svc.name + '</strong></td></tr>' +
                    '<tr><td>版本</td><td><code>' + (svc.version || '-') + '</code></td></tr>' +
                    '<tr><td>ID</td><td><code class="small">' + (svc.id || '-') + '</code></td></tr>' +
                    '<tr><td>状态</td><td><span class="status-badge status-running"><span class="status-dot"></span>' + (svc.status || 'active') + '</span></td></tr></table></div>' +
                    '<div class="col-md-6"><h6 class="text-muted">元数据</h6><table class="table table-sm">' + metadata + '</table></div></div>' +
                    '<h6 class="text-muted mt-3">端点列表</h6><table class="table table-sm"><thead><tr><th>类型</th><th>地址</th><th>路径</th></tr></thead><tbody>' + (endpoints || '<tr><td colspan="3" class="text-muted">无端点</td></tr>') + '</tbody></table>' +
                    '<h6 class="text-muted mt-3">访问示例</h6><pre>curl http://localhost:8888/' + svc.name + '/\ncurl http://localhost:6066/' + svc.name + '/debug/pprof/</pre>';
            } catch (err) {
                detail.innerHTML = '<div class="alert alert-danger">加载失败: ' + err.message + '</div>';
            }
        }
    </script>
</body>
</html>`
}

// getAgentDashboardHTML 返回 Agent 仪表盘 HTML
func getAgentDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel Agent 状态</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        :root {
            --primary-color: #6366f1;
            --success-color: #22c55e;
            --warning-color: #f59e0b;
            --danger-color: #ef4444;
            --bg-dark: #1e1e2e;
            --bg-card: #282a36;
            --text-primary: #f8f8f2;
            --text-secondary: #6272a4;
        }
        body {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            color: var(--text-primary);
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
        }
        .navbar {
            background: rgba(30, 30, 46, 0.95) !important;
            backdrop-filter: blur(10px);
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        .navbar-brand { font-weight: 700; font-size: 1.5rem; }
        .card {
            background: var(--bg-card);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
        }
        .card-header {
            background: transparent;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            font-weight: 600;
        }
        .stat-card {
            background: linear-gradient(135deg, var(--bg-card) 0%, rgba(99, 102, 241, 0.1) 100%);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .stat-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px -5px rgba(99, 102, 241, 0.3);
        }
        .stat-value {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #22c55e, #10b981);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .stat-value.disconnected {
            background: linear-gradient(135deg, #ef4444, #dc2626);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .stat-label {
            color: var(--text-secondary);
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.875rem;
            font-weight: 500;
        }
        .status-connected { background: rgba(34, 197, 94, 0.2); color: var(--success-color); }
        .status-disconnected { background: rgba(239, 68, 68, 0.2); color: var(--danger-color); }
        .status-connecting { background: rgba(245, 158, 11, 0.2); color: var(--warning-color); }
        .status-reconnecting { background: rgba(245, 158, 11, 0.2); color: var(--warning-color); }
        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        .status-connected .status-dot { background: var(--success-color); }
        .status-disconnected .status-dot { background: var(--danger-color); }
        .status-connecting .status-dot { background: var(--warning-color); }
        .status-reconnecting .status-dot { background: var(--warning-color); }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        .endpoint-badge {
            display: inline-block;
            padding: 4px 8px;
            margin: 2px;
            border-radius: 6px;
            font-size: 0.75rem;
            font-weight: 500;
        }
        .endpoint-http { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
        .endpoint-grpc { background: rgba(99, 102, 241, 0.2); color: #6366f1; }
        .endpoint-debug { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
        .refresh-btn { transition: transform 0.3s; }
        .refresh-btn.spinning { animation: spin 1s linear infinite; }
        @keyframes spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }
        pre {
            background: var(--bg-dark);
            border-radius: 8px;
            padding: 16px;
            color: #f8f8f2;
            font-size: 0.875rem;
        }
        .table { color: var(--text-primary); }
        .table thead th {
            background: rgba(99, 102, 241, 0.1);
            border-bottom: 2px solid rgba(99, 102, 241, 0.3);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
        }
        .table tbody tr {
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }
        .text-muted { color: var(--text-secondary) !important; }
        .last-update { font-size: 0.75rem; color: var(--text-secondary); }
        .arch-diagram {
            background: var(--bg-dark);
            border-radius: 12px;
            padding: 20px;
            font-family: 'Fira Code', monospace;
            font-size: 0.75rem;
            line-height: 1.4;
        }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark sticky-top">
        <div class="container-fluid">
            <a class="navbar-brand" href="#">
                <i class="bi bi-arrow-left-right me-2"></i>Tunnel Agent
            </a>
            <div class="d-flex align-items-center">
                <span class="last-update me-3">最后更新: <span id="lastUpdate">-</span></span>
                <button class="btn btn-outline-light btn-sm" onclick="refreshData()">
                    <i class="bi bi-arrow-clockwise refresh-btn" id="refreshIcon"></i> 刷新
                </button>
            </div>
        </div>
    </nav>

    <div class="container-fluid py-4">
        <div class="row g-4 mb-4">
            <div class="col-md-4">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="connectionStatus">检测中</div>
                        <div class="stat-label">连接状态</div>
                    </div>
                </div>
            </div>
            <div class="col-md-4">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <div class="stat-value" id="endpointCount" style="background: linear-gradient(135deg, #6366f1, #8b5cf6); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">0</div>
                        <div class="stat-label">暴露端点</div>
                    </div>
                </div>
            </div>
            <div class="col-md-4">
                <div class="card stat-card h-100">
                    <div class="card-body text-center">
                        <span class="status-badge status-connected" id="agentStatusBadge">
                            <span class="status-dot"></span>
                            <span id="agentStatusText">检测中...</span>
                        </span>
                        <div class="stat-label mt-2">Agent 状态</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row g-4">
            <div class="col-lg-5">
                <div class="card h-100">
                    <div class="card-header"><i class="bi bi-diagram-3 me-2"></i>连接架构</div>
                    <div class="card-body">
                        <div class="arch-diagram">
                            <pre class="mb-0" style="background: transparent; padding: 0;">
┌─────────────────────────┐
│     本服务 (Agent)       │
│  ┌─────┐ ┌─────┐ ┌────┐ │
│  │HTTP │ │gRPC │ │DBG │ │
│  └──┬──┘ └──┬──┘ └─┬──┘ │
└─────┼───────┼──────┼────┘
      └───────┼──────┘
              │ 主动连接 ↓
              ▼
┌─────────────────────────┐
│   Gateway (远端)        │
│      :7007 Tunnel       │
│  ┌─────┐┌─────┐┌─────┐  │
│  │HTTP ││gRPC ││Debug│  │
│  │8888 ││9999 ││6066 │  │
│  └─────┘└─────┘└─────┘  │
└─────────────────────────┘
              │
              ▼
         外部访问</pre>
                        </div>
                        <div class="mt-3">
                            <small class="text-muted">
                                <i class="bi bi-info-circle me-1"></i>
                                Agent 主动连接 Gateway，建立反向隧道，外部通过 Gateway 访问本服务
                            </small>
                        </div>
                    </div>
                </div>
            </div>

            <div class="col-lg-7">
                <div class="card h-100">
                    <div class="card-header"><i class="bi bi-hdd-network me-2"></i>暴露的端点</div>
                    <div class="card-body p-0">
                        <div class="table-responsive">
                            <table class="table table-hover mb-0">
                                <thead>
                                    <tr>
                                        <th>类型</th>
                                        <th>本地地址</th>
                                        <th>路径前缀</th>
                                        <th>Gateway 访问</th>
                                    </tr>
                                </thead>
                                <tbody id="endpointsTable">
                                    <tr>
                                        <td colspan="4" class="text-center text-muted py-4">
                                            <i class="bi bi-hourglass-split me-2"></i>加载中...
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row g-4 mt-2">
            <div class="col-12">
                <div class="card">
                    <div class="card-header"><i class="bi bi-info-circle me-2"></i>连接信息</div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">Gateway 地址</h6>
                                <pre class="mb-0" id="gatewayAddr">-</pre>
                            </div>
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">服务名称</h6>
                                <pre class="mb-0" id="serviceName">-</pre>
                            </div>
                            <div class="col-md-4">
                                <h6 class="text-muted mb-2">服务版本</h6>
                                <pre class="mb-0" id="serviceVersion">-</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let refreshInterval;

        document.addEventListener('DOMContentLoaded', () => {
            refreshData();
            refreshInterval = setInterval(refreshData, 5000);
        });

        async function refreshData() {
            const icon = document.getElementById('refreshIcon');
            icon.classList.add('spinning');
            try {
                await fetchStatus();
                document.getElementById('lastUpdate').textContent = new Date().toLocaleTimeString();
            } catch (err) {
                console.error('Failed to refresh:', err);
            } finally {
                setTimeout(() => icon.classList.remove('spinning'), 500);
            }
        }

        async function fetchStatus() {
            try {
                const res = await fetch('/debug/tunnel/api/status');
                const data = await res.json();
                
                if (data.agent) {
                    const status = data.agent.status || 'disconnected';
                    const statusEl = document.getElementById('agentStatusBadge');
                    const textEl = document.getElementById('agentStatusText');
                    const connEl = document.getElementById('connectionStatus');
                    
                    textEl.textContent = status;
                    statusEl.className = 'status-badge status-' + status;
                    
                    if (status === 'connected') {
                        connEl.textContent = '已连接';
                        connEl.classList.remove('disconnected');
                    } else {
                        connEl.textContent = status === 'connecting' ? '连接中' : 
                                            status === 'reconnecting' ? '重连中' : '未连接';
                        connEl.classList.add('disconnected');
                    }
                    
                    // 更新连接信息
                    if (data.agent.gateway_addr) {
                        document.getElementById('gatewayAddr').textContent = data.agent.gateway_addr;
                    }
                    if (data.agent.service_name) {
                        document.getElementById('serviceName').textContent = data.agent.service_name;
                    }
                    if (data.agent.service_version) {
                        document.getElementById('serviceVersion').textContent = data.agent.service_version;
                    }
                    
                    // 更新端点列表
                    const endpoints = data.agent.endpoints || [];
                    document.getElementById('endpointCount').textContent = endpoints.length;
                    
                    const tbody = document.getElementById('endpointsTable');
                    if (endpoints.length === 0) {
                        tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted py-4"><i class="bi bi-inbox me-2"></i>暂无端点</td></tr>';
                    } else {
                        tbody.innerHTML = endpoints.map(ep => {
                            const gwPort = ep.type === 'http' ? '8888' : ep.type === 'grpc' ? '9999' : '6066';
                            const svcName = data.agent.service_name || 'service';
                            return '<tr>' +
                                '<td><span class="endpoint-badge endpoint-' + ep.type + '">' + ep.type + '</span></td>' +
                                '<td><code>' + ep.address + '</code></td>' +
                                '<td>' + (ep.path || '/') + '</td>' +
                                '<td><code>localhost:' + gwPort + '/' + svcName + ep.path + '</code></td>' +
                                '</tr>';
                        }).join('');
                    }
                }
            } catch (err) { 
                console.error(err); 
            }
        }
    </script>
</body>
</html>`
}

// getEmptyDashboardHTML 返回空状态 HTML（既没有 Gateway 也没有 Agent）
func getEmptyDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tunnel Debug</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.0/font/bootstrap-icons.css" rel="stylesheet">
    <style>
        body {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            color: #f8f8f2;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .empty-state {
            text-align: center;
            padding: 60px;
        }
        .empty-state i {
            font-size: 5rem;
            color: #6272a4;
            margin-bottom: 20px;
        }
        .empty-state h2 {
            color: #f8f8f2;
            margin-bottom: 10px;
        }
        .empty-state p {
            color: #6272a4;
        }
    </style>
</head>
<body>
    <div class="empty-state">
        <i class="bi bi-hdd-network"></i>
        <h2>Tunnel 未配置</h2>
        <p>当前服务未配置 Gateway 或 Agent</p>
        <p class="mt-3"><small>请在代码中调用 <code>tunneldebug.SetGateway()</code> 或 <code>tunneldebug.SetAgent()</code></small></p>
    </div>
</body>
</html>`
}
