package cmux

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/soheilhy/cmux"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	// https://github.com/shaxbee/go-wsproxy
	"go.etcd.io/etcd/client/pkg/v3/transport"
	_ "go.etcd.io/etcd/client/v3/naming/resolver"
	"go.etcd.io/etcd/pkg/v3/httputil"
)

func init() {
	cli, cerr := clientv3.NewFromURL("http://localhost:2379")
	etcdResolver, err := resolver.NewBuilder(cli)
	conn, gerr := grpc.Dial("etcd:///foo/bar/my-service", grpc.WithResolvers(etcdResolver))

	wsproxy.WebsocketProxy(
		gwmux,
		wsproxy.WithRequestMutator(
			// Default to the POST method for streams
			func(_ *http.Request, outgoing *http.Request) *http.Request {
				outgoing.Method = "POST"
				return outgoing
			},
		),
	)

	host := httputil.GetHostname(req)

	m := cmux.New(sctx.l)
	grpcl := m.Match(cmux.HTTP2())
	go func() { errHandler(gs.Serve(grpcl)) }()

	httpl := m.Match(cmux.HTTP1())
	go func() { errHandler(srvhttp.Serve(httpl)) }()

	var tlsl net.Listener
	tlsl, err = transport.NewTLSListener(m.Match(cmux.Any()), tlsinfo)
	if err != nil {
		return err
	}

	m.Serve()
}

func configureHttpServer(srv *http.Server, cfg config.ServerConfig) error {
	// todo (ahrtr): should we support configuring other parameters in the future as well?
	return http2.ConfigureServer(srv, &http2.Server{
		MaxConcurrentStreams: cfg.MaxConcurrentStreams,
	})
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Given in gRPC docs.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// GracefulClose drains http.Response.Body until it hits EOF
// and closes it. This prevents TCP/TLS connections from closing,
// therefore available for reuse.
// Borrowed from golang/net/context/ctxhttp/cancelreq.go.
func GracefulClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

// GetHostname returns the hostname from request Host field.
// It returns empty string, if Host field contains invalid
// value (e.g. "localhost:::" with too many colons).
func GetHostname(req *http.Request) string {
	if req == nil {
		return ""
	}

	h, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		return req.Host
	}
	return h
}

func mustListenCMux(lg *zap.Logger, tlsinfo *transport.TLSInfo) cmux.CMux {
	l, err := net.Listen("tcp", grpcProxyListenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if l, err = transport.NewKeepAliveListener(l, "tcp", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if tlsinfo != nil {
		tlsinfo.CRLFile = grpcProxyListenCRL
		if l, err = transport.NewTLSListener(l, tlsinfo); err != nil {
			lg.Fatal("failed to create TLS listener", zap.Error(err))
		}
	}

	lg.Info("listening for gRPC proxy client requests", zap.String("address", grpcProxyListenAddr))
	return cmux.New(l)
}
