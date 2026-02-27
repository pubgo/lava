package devproxycmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/pubgo/funk/v2/log"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	DNS struct {
		Port int `json:"port" yaml:"port"`
	} `json:"dns" yaml:"dns"`
	HTTP struct {
		Port int `json:"port" yaml:"port"`
	} `json:"http" yaml:"http"`
	Routes []Route `json:"routes" yaml:"routes"`
}

// Route 路由配置
type Route struct {
	Pattern string `json:"pattern" yaml:"pattern"`
	Target  string `json:"target" yaml:"target"`
	Path    string `json:"path" yaml:"path"`
}

// 默认配置
var defaultConfig = Config{
	DNS: struct {
		Port int `json:"port" yaml:"port"`
	}{
		Port: 5353,
	},
	HTTP: struct {
		Port int `json:"port" yaml:"port"`
	}{
		Port: 8080,
	},
	Routes: []Route{
		{
			Pattern: "app",
			Target:  "localhost:3000",
		},
		{
			Pattern: "debug.*",
			Target:  "localhost:8080",
			Path:    "/debug",
		},
		{
			Pattern: "api.*",
			Target:  "localhost:8080",
			Path:    "/{wildcard}",
		},
	},
}

// 全局配置
var config Config

// StartDevProxy 启动开发代理
func StartDevProxy(ctx context.Context) error {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		// 使用默认配置
		config = defaultConfig
	}

	// 启动DNS服务
	go func() {
		if err := startDNSServer(); err != nil {
			log.Error().Err(err).Msg("Failed to start DNS server")
		}
	}()

	// 启动HTTP代理服务
	go func() {
		if err := startHTTPServer(); err != nil {
			log.Error().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	log.Info().Msg("DevProxy started successfully")
	log.Info().Msgf("DNS server listening on port %d", config.DNS.Port)
	log.Info().Msgf("HTTP proxy listening on port %d", config.HTTP.Port)

	// 等待终止信号
	<-ctx.Done()
	log.Info().Msg("DevProxy shutting down")
	return nil
}

// startDNSServer 启动DNS服务器
func startDNSServer() error {
	// 创建DNS服务器
	server := &dns.Server{
		Addr: fmt.Sprintf(":%d", config.DNS.Port),
		Net:  "udp",
	}

	// 注册处理函数，处理所有lava域名
	dns.HandleFunc("lava.", handleDNSRequest)

	// 启动服务器
	log.Info().Msgf("Starting DNS server on port %d", config.DNS.Port)
	return server.ListenAndServe()
}

// handleDNSRequest 处理DNS请求
func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Authoritative = true

	// 处理A记录查询
	for _, q := range r.Question {
		if q.Qtype == dns.TypeA {
			// 所有lava域名都解析到127.0.0.1
			record := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP("127.0.0.1"),
			}
			m.Answer = append(m.Answer, record)
		}
	}

	// 发送响应
	_ = w.WriteMsg(m)
}

// startHTTPServer 启动HTTP代理服务器
func startHTTPServer() error {
	// 创建路由器
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 代理处理
	mux.HandleFunc("/", handleHTTPRequest)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTP.Port),
		Handler: mux,
	}

	// 启动服务器
	log.Info().Msgf("Starting HTTP proxy on %s", server.Addr)
	return server.ListenAndServe()
}

// handleHTTPRequest 处理HTTP请求
func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	// 获取主机名
	host := r.Host
	log.Debug().Str("host", host).Msg("Received request")

	// 移除.lava后缀
	subdomain := strings.TrimSuffix(host, ".lava")

	// 查找匹配的路由
	route, wildcard := matchRoute(subdomain)
	if route == nil {
		http.Error(w, "No route found", http.StatusNotFound)
		return
	}

	// 构建目标URL
	targetURL := buildTargetURL(route, wildcard, r.URL.Path, r.URL.RawQuery)
	log.Debug().Str("target", targetURL).Msg("Forwarding request")

	// 转发所有请求（包括WebSocket）
	forwardRequest(w, r, targetURL)
}

// matchRoute 匹配路由
func matchRoute(subdomain string) (*Route, string) {
	// 精确匹配
	for i := range config.Routes {
		route := &config.Routes[i]
		if route.Pattern == subdomain {
			return route, ""
		}
	}

	// 通配符匹配（从最长到最短）
	parts := strings.Split(subdomain, ".")
	for i := 0; i < len(parts); i++ {
		pattern := strings.Join(parts[i:], ".") + ".*"
		for j := range config.Routes {
			route := &config.Routes[j]
			if route.Pattern == pattern {
				wildcard := strings.Join(parts[:i], ".")
				return route, wildcard
			}
		}
	}

	// 检查是否有默认路由
	for i := range config.Routes {
		route := &config.Routes[i]
		if route.Pattern == "*" {
			return route, subdomain
		}
	}

	return nil, ""
}

// buildTargetURL 构建目标URL
func buildTargetURL(route *Route, wildcard, path, query string) string {
	// 构建路径
	targetPath := route.Path
	if targetPath == "" {
		targetPath = path
	} else {
		// 替换占位符
		targetPath = strings.ReplaceAll(targetPath, "{wildcard}", wildcard)
	}

	// 确保路径格式正确
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	// 处理路径，避免重复的路径前缀
	if path != "/" {
		// 移除路径中可能存在的重复前缀
		trimmedPath := strings.TrimPrefix(path, targetPath)
		if trimmedPath == path {
			// 如果没有重复前缀，直接添加
			if strings.HasSuffix(targetPath, "/") {
				targetPath += strings.TrimPrefix(path, "/")
			} else {
				targetPath += path
			}
		} else {
			// 如果有重复前缀，使用修剪后的路径
			targetPath += trimmedPath
		}
	}

	// 构建完整URL
	targetURL := "http://" + route.Target + targetPath
	if query != "" {
		targetURL += "?" + query
	}

	return targetURL
}

// forwardRequest 转发普通HTTP请求
func forwardRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Error().Err(err).Str("target", targetURL).Msg("Failed to parse target URL")
		http.Error(w, "Failed to parse target URL", http.StatusBadGateway)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().Err(err).Str("target", targetURL).Msg("Failed to forward request")
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
	}

	// 执行代理
	proxy.ServeHTTP(w, r)
}

// loadConfig 加载配置
func loadConfig() error {
	// 尝试从多个位置加载配置
	configPaths := []string{
		".devproxy.json",
		".devproxy.yaml",
		".devproxy.yml",
		"~/.devproxy.json",
		"~/.devproxy.yaml",
		"~/.devproxy.yml",
	}

	for _, path := range configPaths {
		// 处理~路径
		if strings.HasPrefix(path, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			path = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}

		// 检查文件是否存在
		if _, err := os.Stat(path); err == nil {
			// 读取配置文件
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			// 根据文件扩展名解析配置
			ext := filepath.Ext(path)
			switch ext {
			case ".json":
				if err := json.Unmarshal(data, &config); err != nil {
					return fmt.Errorf("failed to parse JSON config: %w", err)
				}
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(data, &config); err != nil {
					return fmt.Errorf("failed to parse YAML config: %w", err)
				}
			default:
				return fmt.Errorf("unsupported config file format: %s", ext)
			}

			log.Info().Str("path", path).Msg("Loaded config from")
			return nil
		}
	}

	// 如果没有找到配置文件，使用默认配置
	config = defaultConfig
	log.Info().Msg("Using default config")
	return nil
}
