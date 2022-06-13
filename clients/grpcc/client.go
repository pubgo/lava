package grpcc

import (
	"context"
	"fmt"
	middleware2 "github.com/pubgo/lava/service"
	"net"
	"strings"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

var _ grpc.ClientConnInterface = (*Client)(nil)

func NewClient(srv string, cfg *grpcc_config.Cfg) *Client {
	return &Client{srv: srv, cfg: cfg}
}

type Client struct {
	cfg  *grpcc_config.Cfg
	mu   sync.Mutex
	conn grpc.ClientConnInterface
	srv  string
}

func (t *Client) createConn(srv string, cfg *grpcc_config.Cfg) (grpc.ClientConnInterface, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DialTimeout)
	defer cancel()

	var middlewares []middleware2.Middleware
	for _, m := range t.cfg.Middlewares {
		middlewares = append(middlewares, m)
	}

	// 加载全局middleware

	addr := t.buildTarget(srv, cfg)
	conn, err := grpc.DialContext(ctx, addr, append(cfg.Client.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(middlewares)),
		grpc.WithChainStreamInterceptor(streamInterceptor(middlewares)))...)

	logging.L().Info("grpc client init", zap.String(logkey.Service, srv))
	logutil.Pretty(err)
	return conn, xerror.WrapF(err, "grpc dial failed, target=>%s", addr)
}

func (t *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var conn, err = t.Get()
	if err != nil {
		return xerror.Wrap(err, method, args)
	}

	return xerror.Wrap(conn.Invoke(ctx, method, args, reply, opts...), method)
}

func (t *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn, err = t.Get()
	if err != nil {
		return nil, xerror.Wrap(err, method)
	}

	var c, err1 = conn.NewStream(ctx, desc, method, opts...)
	return c, xerror.Wrap(err1, method)
}

// Get new grpc Client
func (t *Client) Get() (_ grpc.ClientConnInterface, gErr error) {
	defer xerror.Recovery(func(err xerror.XErr) {
		gErr = err

		logutil.Pretty(t)
		err.DebugPrint()
	})

	if t.conn != nil {
		return t.conn, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双检, 避免多次创建
	if t.conn != nil {
		return t.conn, nil
	}

	var cfg = t.cfg
	if t.cfg == nil {
		t.cfg = grpcc_config.DefaultCfg()
	} else {
		xerror.Panic(merge.Copy(&cfg, grpcc_config.DefaultCfg()))
	}

	conn, err := t.createConn(t.srv, cfg)
	if err != nil {
		return nil, err
	}

	t.conn = conn
	return t.conn, nil
}

func (t *Client) buildTarget(service string, cfg *grpcc_config.Cfg) string {
	var addr = service
	if cfg.Addr != "" {
		addr = cfg.Addr
	}

	if cfg.Registry == "" {
		cfg.Registry = "mdns"
	}

	// 127.0.0.1,127.0.0.1,127.0.0.1;127.0.0.1
	var host = extractHostFromHostPort(addr)
	var scheme = grpcc_resolver.DiscovScheme

	if strings.Contains(addr, ",") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = grpcc_resolver.DirectScheme
	}

	if strings.HasPrefix(service, "k8s://") {
		scheme = grpcc_resolver.K8sScheme
	}

	switch scheme {
	case grpcc_resolver.DiscovScheme:
		return grpcc_resolver.BuildDiscovTarget(addr, cfg.Registry)
	case grpcc_resolver.DirectScheme:
		return grpcc_resolver.BuildDirectTarget(addr)
	case grpcc_resolver.K8sScheme:
		return fmt.Sprintf("dns:///%s", addr)
	default:
		panic("schema is unknown")
	}
}

func extractHostFromHostPort(ep string) string {
	host, _, err := net.SplitHostPort(ep)
	if err != nil {
		return ep
	}
	return host
}
