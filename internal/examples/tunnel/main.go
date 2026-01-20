package main

import (
	"context"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result/resultchecker"

	"github.com/pubgo/lava/v2/cmds/configcmd"
	"github.com/pubgo/lava/v2/cmds/envcmd"
	"github.com/pubgo/lava/v2/core/debug/tunneldebug"
	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/logging"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
	"github.com/pubgo/lava/v2/servers/https"
)

// Config 配置结构
type Config struct {
	metrics.MetricConfigLoader   `yaml:",inline"`
	logging.LogConfigLoader      `yaml:",inline"`
	https.HttpServerConfigLoader `yaml:",inline"`

	// Tunnel Gateway 配置
	Tunnel *TunnelConfig `yaml:"tunnel"`
}

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

// NewTunnelGatewayService 创建 Tunnel Gateway 服务
func NewTunnelGatewayService(cfg *Config) supervisor.Service {
	gateway, err := tunnel.NewGatewayBuilder().
		WithListenAddr(cfg.Tunnel.ListenAddr).
		WithHTTPPort(cfg.Tunnel.HTTPPort).
		WithGRPCPort(cfg.Tunnel.GRPCPort).
		WithDebugPort(cfg.Tunnel.DebugPort).
		Build()
	if err != nil {
		panic(err)
	}

	// 注册到 debug 界面
	tunneldebug.SetGateway(gateway)

	return &tunnelGatewayService{gateway: gateway}
}

func main() {
	defer recovery.Exit()

	version.SetVersion("v1.0.0")
	version.SetProject("tunnel-gateway")
	config.SetConfigPath("internal/configs/tunnel.yaml")
	resultchecker.RegisterErrCheck(log.RecordErr())
	log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Fields) bool {
		//if lvl == zerolog.DebugLevel {
		//	return false
		//}
		return true
	})
	//debugs.SetEnabled()
	env.LoadFiles(".env").Must()

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(envcmd.New)
	builder.Provide(configcmd.New[Config])
	builder.Provide(NewTunnelGatewayService)

	lavabuilder.Run(builder)
}
