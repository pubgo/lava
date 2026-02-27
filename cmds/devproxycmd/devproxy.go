package devproxycmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/miekg/dns"
	"github.com/pubgo/funk/v2/closer"
	"github.com/pubgo/funk/v2/log"
	"github.com/valyala/fasthttp"
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
	// 创建Fiber应用
	app := fiber.New()

	// 健康检查端点
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 代理处理
	app.All("/*", handleHTTPRequest)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.HTTP.Port)
	log.Info().Msgf("Starting HTTP proxy on %s", addr)
	return app.Listen(addr)
}

// handleHTTPRequest 处理HTTP请求
func handleHTTPRequest(c fiber.Ctx) error {
	// 获取主机名
	host := c.Hostname()
	log.Debug().Str("host", host).Msg("Received request")

	// 移除.lava后缀
	subdomain := strings.TrimSuffix(host, ".lava")

	// 查找匹配的路由
	route, wildcard := matchRoute(subdomain)
	if route == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No route found",
			"host":  host,
		})
	}

	// 构建目标URL
	targetURL := buildTargetURL(route, wildcard, c.Path(), c.Request().URI().QueryArgs().String())
	log.Debug().Str("target", targetURL).Msg("Forwarding request")

	// 检测是否为websocket连接
	if c.Get("Upgrade") == "websocket" {
		return forwardWebSocket(c, targetURL)
	}

	// 转发普通HTTP请求
	return forwardRequest(c, targetURL)
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
func forwardRequest(c fiber.Ctx, targetURL string) error {
	// 创建fasthttp客户端
	client := &fasthttp.Client{
		MaxConnsPerHost: 100,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
	}

	// 创建目标请求
	var req fasthttp.Request
	req.SetRequestURI(targetURL)
	req.Header.SetMethod(c.Method())

	// 复制请求头
	for key, value := range c.Request().Header.All() {
		req.Header.SetBytesKV(key, value)
	}

	// 复制请求体
	req.SetBody(c.Request().Body())

	// 发送请求并获取响应
	var resp fasthttp.Response
	if err := client.Do(&req, &resp); err != nil {
		log.Error().Err(err).Str("target", targetURL).Msg("Failed to forward request")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Failed to forward request",
			"message": err.Error(),
			"target":  targetURL,
		})
	}

	// 复制响应状态码
	c.Status(resp.StatusCode())

	// 复制响应头

	for key, value := range resp.Header.All() {
		c.Set(string(key), string(value))
	}

	// 复制响应体
	return c.Send(resp.Body())
}

// forwardWebSocket 处理websocket连接
func forwardWebSocket(c fiber.Ctx, targetURL string) error {
	// 解析目标URL
	u, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	// 将http://或https://转换为ws://或wss://
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}
	wsURL := wsScheme + "://" + u.Host + u.Path
	if u.RawQuery != "" {
		wsURL += "?" + u.RawQuery
	}

	// 升级当前连接为WebSocket
	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true // 允许所有来源的连接
		},
	}

	return upgrader.Upgrade(c.RequestCtx(), func(conn *websocket.Conn) {
		defer closer.SafeClose(conn)

		// 连接到目标WebSocket服务器
		targetConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			log.Error().Err(err).Str("url", wsURL).Str("target", targetURL).Msg("Failed to connect to target WebSocket server")
			return
		}
		defer closer.SafeClose(targetConn)

		// 双向数据转发
		done := make(chan struct{})

		// 从客户端读取数据并发送到目标服务器
		go func() {
			defer close(done)
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				if err := targetConn.WriteMessage(messageType, message); err != nil {
					break
				}
			}
		}()

		// 从目标服务器读取数据并发送到客户端
		go func() {
			defer close(done)
			for {
				messageType, message, err := targetConn.ReadMessage()
				if err != nil {
					break
				}
				if err := conn.WriteMessage(messageType, message); err != nil {
					break
				}
			}
		}()

		<-done
	})
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
