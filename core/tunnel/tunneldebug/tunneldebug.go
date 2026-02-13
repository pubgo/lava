// Package tunneldebug 提供 Tunnel Gateway 的 Web 管理界面
package tunneldebug

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/adaptor/v2"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/tunnel"
)

var (
	globalGateway tunnel.Gateway
	globalAgent   tunnel.Agent
	mu            sync.RWMutex
)

// SetGateway 设置全局 Gateway 实例
func SetGateway(gw tunnel.Gateway) {
	mu.Lock()
	defer mu.Unlock()
	globalGateway = gw
}

// SetAgent 设置全局 Agent 实例
func SetAgent(agent tunnel.Agent) {
	mu.Lock()
	defer mu.Unlock()
	globalAgent = agent
}

func init() {
	// 注册路由到 debug app
	debug.Get("/tunnel", adaptor.HTTPHandlerFunc(handleDashboard))
	debug.Get("/tunnel/", adaptor.HTTPHandlerFunc(handleDashboard))
	debug.Get("/tunnel/api/status", adaptor.HTTPHandlerFunc(handleAPIStatus))
	debug.Get("/tunnel/api/services", adaptor.HTTPHandlerFunc(handleAPIServices))
	debug.Get("/tunnel/api/services/:name", adaptor.HTTPHandlerFunc(handleAPIServiceDetail))
	debug.Get("/tunnel/api/stats", adaptor.HTTPHandlerFunc(handleAPIStats))
}

// handleDashboard 渲染主界面
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	gw := globalGateway
	agent := globalAgent
	mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 根据是 Gateway 还是 Agent 选择不同的 HTML
	if gw != nil {
		w.Write([]byte(getGatewayDashboardHTML()))
	} else if agent != nil {
		w.Write([]byte(getAgentDashboardHTML()))
	} else {
		w.Write([]byte(getEmptyDashboardHTML()))
	}
}

// handleAPIStatus 返回状态 JSON
func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	gw := globalGateway
	agent := globalAgent
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	status := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if gw != nil {
		services := gw.Services()
		status["gateway"] = map[string]any{
			"status":        gw.Status().String(),
			"service_count": len(services),
		}
	}

	if agent != nil {
		// 获取 Agent 详细信息
		info := agent.Info()
		status["agent"] = map[string]any{
			"status":          info.Status,
			"gateway_addr":    info.GatewayAddr,
			"service_name":    info.ServiceName,
			"service_version": info.ServiceVersion,
			"endpoints":       info.Endpoints,
		}
	}

	json.NewEncoder(w).Encode(status)
}

// handleAPIServices 返回服务列表
func handleAPIServices(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	gw := globalGateway
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if gw == nil {
		json.NewEncoder(w).Encode(map[string]any{
			"services": []any{},
			"total":    0,
		})
		return
	}

	services := gw.Services()

	// 排序
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	json.NewEncoder(w).Encode(map[string]any{
		"services": services,
		"total":    len(services),
	})
}

// handleAPIServiceDetail 返回服务详情
func handleAPIServiceDetail(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	gw := globalGateway
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if gw == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "gateway not configured"})
		return
	}

	// 从 URL 获取服务名
	name := strings.TrimPrefix(r.URL.Path, "/debug/tunnel/api/services/")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "service name required"})
		return
	}

	svc, err := gw.GetService(name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(svc)
}

// handleAPIStats 返回统计信息
func handleAPIStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	gw := globalGateway
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	stats := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if gw != nil {
		services := gw.Services()

		// 统计端点类型
		endpointStats := map[string]int{
			"http":  0,
			"grpc":  0,
			"debug": 0,
		}

		for _, svc := range services {
			for _, ep := range svc.Endpoints {
				endpointStats[string(ep.Type)]++
			}
		}

		stats["services"] = map[string]any{
			"total":     len(services),
			"endpoints": endpointStats,
		}
	}

	json.NewEncoder(w).Encode(stats)
}
