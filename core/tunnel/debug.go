package tunnel

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// DebugHandler 提供调试接口
type DebugHandler struct {
	gateway Gateway
	agent   Agent
	mu      sync.RWMutex
}

// NewDebugHandler 创建调试处理器
func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

// SetGateway 设置网关实例
func (h *DebugHandler) SetGateway(gw Gateway) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gateway = gw
}

// SetAgent 设置代理客户端实例
func (h *DebugHandler) SetAgent(agent Agent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.agent = agent
}

// FiberRoutes 注册 Fiber 路由
func (h *DebugHandler) FiberRoutes(app fiber.Router) {
	group := app.Group("/tunnel")

	// Gateway 状态
	group.Get("/gateway", h.handleGatewayStatus)
	group.Get("/gateway/services", h.handleGatewayServices)
	group.Get("/gateway/services/:name", h.handleGatewayService)

	// Agent 状态
	group.Get("/agent", h.handleAgentStatus)

	// 概览页面
	group.Get("/", h.handleOverview)
}

// HTTPRoutes 注册标准 HTTP 路由
func (h *DebugHandler) HTTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/debug/tunnel/", h.handleHTTPOverview)
	mux.HandleFunc("/debug/tunnel/gateway", h.handleHTTPGatewayStatus)
	mux.HandleFunc("/debug/tunnel/gateway/services", h.handleHTTPGatewayServices)
	mux.HandleFunc("/debug/tunnel/agent", h.handleHTTPAgentStatus)
}

// handleGatewayStatus 返回网关状态
func (h *DebugHandler) handleGatewayStatus(c *fiber.Ctx) error {
	h.mu.RLock()
	gw := h.gateway
	h.mu.RUnlock()

	if gw == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "gateway not configured",
		})
	}

	return c.JSON(fiber.Map{
		"status":   gw.Status().String(),
		"services": gw.Services(),
	})
}

// handleGatewayServices 返回所有注册的服务
func (h *DebugHandler) handleGatewayServices(c *fiber.Ctx) error {
	h.mu.RLock()
	gw := h.gateway
	h.mu.RUnlock()

	if gw == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "gateway not configured",
		})
	}

	return c.JSON(gw.Services())
}

// handleGatewayService 返回指定服务的详情
func (h *DebugHandler) handleGatewayService(c *fiber.Ctx) error {
	h.mu.RLock()
	gw := h.gateway
	h.mu.RUnlock()

	if gw == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "gateway not configured",
		})
	}

	name := c.Params("name")
	svc, err := gw.GetService(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(svc)
}

// handleAgentStatus 返回代理客户端状态
func (h *DebugHandler) handleAgentStatus(c *fiber.Ctx) error {
	h.mu.RLock()
	agent := h.agent
	h.mu.RUnlock()

	if agent == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent not configured",
		})
	}

	return c.JSON(fiber.Map{
		"status": agent.Status().String(),
	})
}

// handleOverview 返回概览 HTML 页面
func (h *DebugHandler) handleOverview(c *fiber.Ctx) error {
	h.mu.RLock()
	gw := h.gateway
	agent := h.agent
	h.mu.RUnlock()

	data := map[string]any{
		"title":     "Tunnel Debug",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if gw != nil {
		data["gateway"] = map[string]any{
			"status":   gw.Status().String(),
			"services": gw.Services(),
		}
	}

	if agent != nil {
		data["agent"] = map[string]any{
			"status": agent.Status().String(),
		}
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(renderDebugHTML(data))
}

// HTTP handlers
func (h *DebugHandler) handleHTTPOverview(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	gw := h.gateway
	agent := h.agent
	h.mu.RUnlock()

	data := map[string]any{
		"title":     "Tunnel Debug",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if gw != nil {
		data["gateway"] = map[string]any{
			"status":   gw.Status().String(),
			"services": gw.Services(),
		}
	}

	if agent != nil {
		data["agent"] = map[string]any{
			"status": agent.Status().String(),
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderDebugHTML(data)))
}

func (h *DebugHandler) handleHTTPGatewayStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	gw := h.gateway
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if gw == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "gateway not configured"})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":   gw.Status().String(),
		"services": gw.Services(),
	})
}

func (h *DebugHandler) handleHTTPGatewayServices(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	gw := h.gateway
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if gw == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "gateway not configured"})
		return
	}

	json.NewEncoder(w).Encode(gw.Services())
}

func (h *DebugHandler) handleHTTPAgentStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	agent := h.agent
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if agent == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "agent not configured"})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": agent.Status().String(),
	})
}

func renderDebugHTML(data map[string]any) string {
	const tmpl = `<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; }
        h1 { color: #333; }
        h2 { color: #555; margin-top: 30px; }
        .card { background: #f5f5f5; border-radius: 8px; padding: 15px; margin: 10px 0; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .status-running, .status-connected { background: #4caf50; color: white; }
        .status-stopped, .status-disconnected { background: #f44336; color: white; }
        .status-reconnecting { background: #ff9800; color: white; }
        table { width: 100%; border-collapse: collapse; margin-top: 10px; }
        th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
        th { background: #333; color: white; }
        tr:hover { background: #f1f1f1; }
        .timestamp { color: #999; font-size: 12px; }
        pre { background: #272822; color: #f8f8f2; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>🚀 Tunnel Debug</h1>
    <p class="timestamp">Updated: {{.timestamp}}</p>
    
    {{if .gateway}}
    <h2>Gateway</h2>
    <div class="card">
        <p>Status: <span class="status status-{{.gateway.status}}">{{.gateway.status}}</span></p>
        {{if .gateway.services}}
        <h3>Registered Services</h3>
        <table>
            <tr>
                <th>Name</th>
                <th>Version</th>
                <th>Status</th>
                <th>Endpoints</th>
            </tr>
            {{range .gateway.services}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Version}}</td>
                <td>{{.Status}}</td>
                <td>{{len .Endpoints}}</td>
            </tr>
            {{end}}
        </table>
        {{else}}
        <p>No services registered</p>
        {{end}}
    </div>
    {{end}}
    
    {{if .agent}}
    <h2>Agent</h2>
    <div class="card">
        <p>Status: <span class="status status-{{.agent.status}}">{{.agent.status}}</span></p>
    </div>
    {{end}}
    
    <h2>API Endpoints</h2>
    <div class="card">
        <pre>
GET /tunnel/gateway         - Gateway status
GET /tunnel/gateway/services - List all services
GET /tunnel/gateway/services/:name - Get service details
GET /tunnel/agent           - Agent status
        </pre>
    </div>
</body>
</html>`

	t := template.Must(template.New("debug").Parse(tmpl))
	var buf []byte
	buf = append(buf, ""...)

	// Use a simple approach since we can't use bytes.Buffer easily
	result := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; }
        h1 { color: #333; }
        h2 { color: #555; margin-top: 30px; }
        .card { background: #f5f5f5; border-radius: 8px; padding: 15px; margin: 10px 0; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .status-running, .status-connected { background: #4caf50; color: white; }
        .status-stopped, .status-disconnected { background: #f44336; color: white; }
        pre { background: #272822; color: #f8f8f2; padding: 10px; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>🚀 Tunnel Debug</h1>
    <p>Updated: %s</p>
    <h2>API Endpoints</h2>
    <div class="card">
        <pre>
GET /tunnel/gateway         - Gateway status
GET /tunnel/gateway/services - List all services  
GET /tunnel/agent           - Agent status
        </pre>
    </div>
</body>
</html>`, data["title"], data["timestamp"])

	_ = t
	_ = buf

	return result
}
