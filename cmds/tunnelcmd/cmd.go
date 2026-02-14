package tunnelcmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	return &supervisor.Metric{
		Name:   s.Name(),
		Status: supervisor.StatusRunning,
	}
}

// debugServerService 内嵌的 debug 服务器
type debugServerService struct {
	app  *fiber.App
	addr string
	err  error
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
	if err := s.app.Listen(s.addr); err != nil && err != http.ErrServerClosed {
		s.err = err
		return err
	}
	return nil
}

func (s *debugServerService) Metric() *supervisor.Metric {
	return &supervisor.Metric{
		Name:   s.Name(),
		Status: supervisor.StatusRunning,
	}
}

// newDebugServer 创建 debug 服务器
func newDebugServer(addr string) *debugServerService {
	app := fiber.New()

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
			return nil
		},
		Children: []*redant.Command{
			newGatewayCommand(di),
		},
	}
}

func newGatewayCommand(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "gateway",
		Short: cliutil.UsageDesc("tunnel gateway service %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			// 设置默认值（直接从环境变量读取，不依赖配置文件）
			tunnelCfg := &TunnelConfig{
				ListenAddr: ":7007",
				HTTPPort:   8888,
				GRPCPort:   9999,
				DebugPort:  6066,
			}

			// 环境变量覆盖
			if addr := os.Getenv("TUNNEL_LISTEN_ADDR"); addr != "" {
				tunnelCfg.ListenAddr = addr
			}
			if port := os.Getenv("TUNNEL_HTTP_PORT"); port != "" {
				if p, err := parsePort(port); err == nil {
					tunnelCfg.HTTPPort = p
				}
			}
			if port := os.Getenv("TUNNEL_GRPC_PORT"); port != "" {
				if p, err := parsePort(port); err == nil {
					tunnelCfg.GRPCPort = p
				}
			}
			if port := os.Getenv("TUNNEL_DEBUG_PORT"); port != "" {
				if p, err := parsePort(port); err == nil {
					tunnelCfg.DebugPort = p
				}
			}

			// 创建 Gateway
			gateway := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
				ListenAddr: tunnelCfg.ListenAddr,
				HTTPPort:   tunnelCfg.HTTPPort,
				GRPCPort:   tunnelCfg.GRPCPort,
				DebugPort:  tunnelCfg.DebugPort,
			})

			err := gateway.Start(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Gateway start failed")
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
				Str("listen_addr", tunnelCfg.ListenAddr).
				Int("http_port", tunnelCfg.HTTPPort).
				Int("grpc_port", tunnelCfg.GRPCPort).
				Int("debug_port", tunnelCfg.DebugPort).
				Str("admin_addr", debugAddr).
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
