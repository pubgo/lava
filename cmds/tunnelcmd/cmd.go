package tunnelcmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/supervisor"
	supervisordebug "github.com/pubgo/lava/v2/core/supervisor/debug"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunneldebug"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

// TunnelConfig Tunnel Gateway 配置
type TunnelConfig struct {
	// ListenAddr Agent 连接监听地址
	ListenAddr string `yaml:"listen_addr" default:":7007"`
	// HTTPPort HTTP 代理端口
	HTTPPort int `yaml:"http_port" default:"8888"`
	// GRPCPort gRPC 代理端口
	GRPCPort int `yaml:"grpc_port" default:"9999"`
	// DebugPort Debug 代理端口
	DebugPort int `yaml:"debug_port" default:"6066"`
}

// Config 配置结构
type Config struct {
	Tunnel *TunnelConfig `yaml:"tunnel"`
}

// tunnelGatewayService 包装 Gateway 为 supervisor.Service
type tunnelGatewayService struct {
	gateway tunnel.Gateway
	err     error
}

func (s *tunnelGatewayService) Name() string {
	return "tunnel-gateway"
}

func (s *tunnelGatewayService) Error() error {
	return s.err
}

func (s *tunnelGatewayService) String() string {
	return "Tunnel Gateway Service - accepts agent connections and proxies requests"
}

func (s *tunnelGatewayService) Serve(ctx context.Context) error {
	if err := s.gateway.Start(ctx); err != nil {
		s.err = err
		return err
	}

	// 等待上下文取消
	<-ctx.Done()

	return s.gateway.Stop(context.Background())
}

func (s *tunnelGatewayService) Metric() *supervisor.Metric {
	m := &supervisor.Metric{Name: s.Name()}
	if s.err != nil {
		m.Status = supervisor.StatusError
		m.LastError = s.err.Error()
		return m
	}

	switch s.gateway.Status() {
	case tunnel.GatewayStatusRunning:
		m.Status = supervisor.StatusRunning
	case tunnel.GatewayStatusStarting:
		m.Status = supervisor.StatusIdle
	case tunnel.GatewayStatusStopping:
		m.Status = supervisor.StatusStopped
	default:
		m.Status = supervisor.StatusStopped
	}
	return m
}

// debugServerService 内嵌的 debug 服务器
type debugServerService struct {
	app     *fiber.App
	addr    string
	err     error
	running atomic.Bool
}

func (s *debugServerService) Name() string { return "debug-server" }
func (s *debugServerService) Error() error { return s.err }
func (s *debugServerService) String() string {
	return "Debug Server - provides management UI at " + s.addr
}

func (s *debugServerService) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.app.Shutdown()
	}()

	log.Info().Str("addr", s.addr).Msg("Debug server started")
	s.running.Store(true)
	defer s.running.Store(false)

	if err := s.app.Listen(s.addr); err != nil && err != http.ErrServerClosed {
		s.err = err
		return err
	}
	return nil
}

func (s *debugServerService) Metric() *supervisor.Metric {
	m := &supervisor.Metric{Name: s.Name()}
	if s.err != nil {
		m.Status = supervisor.StatusError
		m.LastError = s.err.Error()
		return m
	}
	if s.running.Load() {
		m.Status = supervisor.StatusRunning
	} else {
		m.Status = supervisor.StatusStopped
	}
	return m
}

// newDebugServer 创建 debug 服务器
func newDebugServer(addr string) *debugServerService {
	app := fiber.New()

	adminToken := tunnel.AdminTokenFromEnv()
	if adminToken != "" {
		app.Use(adminAuthMiddleware(adminToken))
	} else if !strings.HasPrefix(addr, "127.0.0.1:") && !strings.HasPrefix(addr, "localhost:") {
		log.Warn().Str("addr", addr).Msg("TUNNEL_ADMIN_TOKEN not set: admin UI is unauthenticated")
	}

	// 挂载 debug 路由
	app.Use("/debug", debug.App())

	// 根路由重定向到 tunnel dashboard
	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/debug/tunnel")
	})

	return &debugServerService{
		app:  app,
		addr: addr,
	}
}

// New 创建 tunnel gateway 命令
func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "tunnel",
		Short: cliutil.UsageDesc("tunnel service %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			fmt.Println("Usage: lava tunnel [command] [arguments]")
			fmt.Println("Available commands:")
			fmt.Println("  gateway   Run tunnel gateway")
			fmt.Println("  agent     Run tunnel agent (set P2P_PEER_ID to enable P2P)")
			return nil
		},
		Children: []*redant.Command{
			newGatewayCommand(di),
			newAgentCommand(di),
		},
	}
}

func newGatewayCommand(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "gateway",
		Short: cliutil.UsageDesc("tunnel gateway service %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			gwCfg := tunnel.GatewayConfigFromEnv()
			// CLI 历史默认与 DefaultGatewayConfig 不同，保持兼容
			if os.Getenv("TUNNEL_HTTP_PORT") == "" {
				gwCfg.HTTPPort = 8888
			}
			if os.Getenv("TUNNEL_GRPC_PORT") == "" {
				gwCfg.GRPCPort = 9999
			}
			if os.Getenv("TUNNEL_DEBUG_PORT") == "" {
				gwCfg.DebugPort = 6066
			}
			gwCfg.Transport = tunnel.TransportYamux
			gwCfg.Normalize()

			gateway := tunnelgateway.NewGateway(&gwCfg)
			authToken := tunnel.AuthTokenFromEnv()
			if authToken == "" {
				log.Warn().Msg("TUNNEL_AUTH_TOKEN not set: any agent may register; HTTP/gRPC/debug proxies remain unauthenticated")
			} else {
				log.Info().Msg("Proxy authentication enabled: clients must send Authorization: Bearer or X-Tunnel-Token")
			}
			if err := tunnel.ConfigureGatewayAuth(gateway, authToken); err != nil {
				return err
			}

			// 注册到 debug 界面
			tunneldebug.SetGateway(gateway)

			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service
			}))

			manager := supervisor.Default(params.LC)
			supervisordebug.Register(manager)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			// 添加 Gateway 服务
			if err := manager.Add(&tunnelGatewayService{gateway: gateway}); err != nil {
				return err
			}

			// 添加 Debug 服务器（管理界面）
			debugAddr := ":6067" // 使用独立端口，避免与 debug proxy 冲突
			if addr := os.Getenv("TUNNEL_ADMIN_ADDR"); addr != "" {
				debugAddr = addr
			}
			if err := manager.Add(newDebugServer(debugAddr)); err != nil {
				return err
			}

			log.Info().
				Str("listen_addr", gwCfg.ListenAddr).
				Int("http_port", gwCfg.HTTPPort).
				Int("grpc_port", gwCfg.GRPCPort).
				Int("debug_port", gwCfg.DebugPort).
				Str("admin_addr", debugAddr).
				Bool("auth_enabled", authToken != "").
				Bool("tls_enabled", gwCfg.TransportOptions != nil && gwCfg.TransportOptions.EnableTLS).
				Msg("Starting Tunnel Gateway")

			return manager.Run(ctx)
		},
	}
}

func parsePort(s string) (int, error) {
	var port int
	_, err := fmt.Sscanf(s, "%d", &port)
	return port, err
}
