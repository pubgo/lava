package golug_grpc

import (
	"context"
	"crypto/tls"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const (
	// DefaultMaxMsgSize define maximum message size that srv can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4

	defaultContentType          = "application/grpc"
	DefaultSleepAfterDeregister = time.Second * 2
)

var (
	/*
		var kasp = keepalive.ServerParameters{
			//MaxConnectionIdle:     30 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
			//MaxConnectionAge:      55 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
			//MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:    10 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout: 2 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
		}
	*/
	kaep = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}
	defaultUnaryTimeout                                             = 10 * time.Second
	defaultStreamTimeout                                            = 10 * time.Second
	registryName                                                    = ""
	defaultRateLimiter     ratelimit.Limiter                        = RateLimit{}
	defaultAuthFunc        grpc_auth.AuthFunc                       = Auth
	defaultRecoveryHandler grpc_recovery.RecoveryHandlerFuncContext = recoveryHandler
	tlsCfg                 *tls.Config
	Metadata               = make(map[string]string)
	Address                string
	Advertise              string
	Id                     string
	Version                string

	// The register expiry time
	RegisterTTL = time.Minute
	// The interval on which to register
	RegisterInterval = time.Second * 30
)

var streamInterceptors = []grpc.StreamServerInterceptor{
	grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
	grpc_opentracing.StreamServerInterceptor(),
	ratelimit.StreamServerInterceptor(defaultRateLimiter),
	grpc_auth.StreamServerInterceptor(defaultAuthFunc)}

var unaryInterceptors = []grpc.UnaryServerInterceptor{
	grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
	grpc_opentracing.UnaryServerInterceptor(),
	ratelimit.UnaryServerInterceptor(defaultRateLimiter),
	grpc_auth.UnaryServerInterceptor(defaultAuthFunc)}

func GetDefaultServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultMaxMsgSize),
		grpc.MaxSendMsgSize(DefaultMaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(kaep),
	}
}

type RateLimit struct{}

func (r RateLimit) Limit() bool {
	return false
}

func Auth(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func recoveryHandler(ctx context.Context, p interface{}) (err error) {
	xlog.Errorf("handler is panic: %v", p)
	return status.Errorf(codes.Internal, "%s", p)
}
