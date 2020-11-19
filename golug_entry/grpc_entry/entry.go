package grpc_entry

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_data"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
)

var _ golug_entry.GrpcEntry = (*grpcEntry)(nil)

type grpcEntry struct {
	golug_entry.Entry
	server                   *grpc.Server
	opts                     golug_entry.Options
	handlers                 []interface{}
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor
}

func (t *grpcEntry) Run() golug_entry.RunEntry { return t }

func (t *grpcEntry) UnWrap(fn interface{}) error { return xerror.Wrap(golug_entry.UnWrap(t, fn)) }

func (t *grpcEntry) Register(ss interface{}) {
	if ss == nil {
		xerror.Panic(xerror.New("[ss] should not be nil"))
	}

	t.handlers = append(t.handlers, ss)
}

func (t *grpcEntry) WithUnaryServer(interceptors ...grpc.UnaryServerInterceptor) {
	t.unaryServerInterceptors = append(t.unaryServerInterceptors, interceptors...)
}

func (t *grpcEntry) WithStreamServer(interceptors ...grpc.StreamServerInterceptor) {
	t.streamServerInterceptors = append(t.streamServerInterceptors, interceptors...)
}

func (t *grpcEntry) Init() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.Entry.Run().Init())

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnknownServiceHandler(g.handler),
		grpc.Creds(credentials.NewTLS(v.(*tls.Config))),
	}

	// 初始化server
	t.server = grpc.NewServer(gopts...)
	return nil
}

func (t *grpcEntry) Start() (err error) {
	defer xerror.RespErr(&err)

	// 初始化routes
	for i := range t.handlers {
		xerror.Panic(register(t.server, t.handlers[i]))
	}

	cancel := xprocess.Go(func(ctx context.Context) (err error) {
		defer xerror.RespErr(&err)

		ts := xerror.PanicErr(net.Listen("tcp", t.Entry.Run().Options().Addr)).(net.Listener)
		log.Infof("Server [grpc] Listening on %s", ts.Addr().String())
		if err := t.server.Serve(ts); err != nil {
			log.Error(err.Error())
		}
		return nil
	})

	xerror.Panic(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { xerror.Panic(cancel()) }))

	return nil
}

func (t *grpcEntry) Stop() (err error) {
	defer xerror.RespErr(&err)
	t.server.GracefulStop()
	log.Infof("Server [grpc] Closed OK")
	return nil
}

func (t *grpcEntry) initCfg() {
	xerror.Panic(golug_config.Decode("server", &t.opts.RestCfg))
}

func (t *grpcEntry) initFlags() {
	xerror.Panic(t.Flags(func(flags *pflag.FlagSet) {
		flags.StringVar(&t.opts.Addr, "http_addr", t.opts.Addr, "the http server address")
		flags.BoolVar(&t.opts.RestCfg.DisableStartupMessage, "disable_startup_message", t.opts.RestCfg.DisableStartupMessage, "print out the http server art and listening address")
	}))
}

func newEntry(name string) *grpcEntry {
	ent := &grpcEntry{
		Entry: golug_entry.New(name),
	}
	ent.initFlags()
	ent.initCfg()
	ent.trace()

	return ent
}

func New(name string) *grpcEntry {
	return newEntry(name)
}

func GetDefaultServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
			TimeoutUnaryServerInterceptor(defaultUnaryTimeout),
			grpc_opentracing.UnaryServerInterceptor(),
			ratelimit.UnaryServerInterceptor(defaultRateLimiter),
			grpc_auth.UnaryServerInterceptor(defaultAuthFunc),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(defaultRecoveryHandler)),
			TimeoutStreamServerInterceptor(defaultStreamTimeout),
			grpc_opentracing.StreamServerInterceptor(),
			ratelimit.StreamServerInterceptor(defaultRateLimiter),
			grpc_auth.StreamServerInterceptor(defaultAuthFunc),
		)),
	}

}

func TimeoutUnaryServerInterceptor(t time.Duration) grpc.UnaryServerInterceptor {
	defaultTimeOut := t
	if t := os.Getenv("GRPC_UNARY_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeOut = time.Duration(s) * time.Second
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := ctx.Deadline(); !ok { //if ok is true, it is set by header grpc-timeout from client
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
			defer cancel()
		}

		// create a done channel to tell the request it's done
		done := make(chan struct{})

		// here you put the actual work needed for the request
		// and then send the doneChan with the status and body
		// to finish the request by writing the response
		var res interface{}
		var err error

		go func() {
			defer func() {
				if c := recover(); c != nil {
					log.Errorf("response request panic: %v", c)
					err = status.Errorf(codes.Internal, "response request panic: %v", c)
				}
				close(done)
			}()
			res, err = handler(ctx, req)
		}()

		// non-blocking select on two channels see if the request
		// times out or finishes
		select {

		// if the context is done it timed out or was canceled
		// so don't return anything
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "handler timeout")

		// if the request finished then finish the request by
		// writing the response
		case <-done:
			return res, err
		}
	}
}

func TimeoutStreamServerInterceptor(defaultTimeOut time.Duration) grpc.StreamServerInterceptor {
	if t := os.Getenv("GRPC_STREAM_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeOut = time.Duration(s) * time.Second
		}
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := stream.Context()
		if _, ok := ctx.Deadline(); !ok { //if ok is true, it is set by header grpc-timeout from client
			if defaultTimeOut == 0 {
				return handler(srv, stream)
			}
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
			defer cancel()
		}
		// create a done channel to tell the request it's done
		done := make(chan struct{})

		go func() {
			defer func() {
				if c := recover(); c != nil {
					log.Errorf("response request panic: %v", c)
					err = status.Errorf(codes.Internal, "response request panic: %v", c)
				}
				close(done)
			}()
			err = handler(srv, stream)
		}()

		// non-blocking select on two channels see if the request
		// times out or finishes
		select {

		// if the context is done it timed out or was canceled
		// so don't return anything
		case <-ctx.Done():
			return status.Errorf(codes.DeadlineExceeded, "handler timeout")

		// if the request finished then finish the request by
		// writing the response
		case <-done:
			return err
		}
	}
}

// middleware for grpc unary calls
var defaultUnaryInterceptor = grpc_middleware.ChainUnaryClient(
	grpc_opentracing.UnaryClientInterceptor(),
)

// middleware for grpc stream calls
var defaultStreamInterceptor = grpc_middleware.ChainStreamClient(grpc_opentracing.StreamClientInterceptor())

func init() {
	timeoutCtx, _ := context.WithTimeout(context.Background(), 0)
	conn, err := grpc.DialContext(timeoutCtx, golug_config.Project, grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor))
}

func register(server *grpc.Server, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	if handler == nil {
		return xerror.New("[handler] should not be nil")
	}

	if server == nil {
		return xerror.New("[server] should not be nil")
	}

	var vRegister reflect.Value
	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for _, v := range golug_data.List() {
		v1 := reflect.TypeOf(v)
		if v1.NumIn() < 2 {
			continue
		}

		if hd.Implements(v1.In(1)) {
			vRegister = reflect.ValueOf(v)
			break
		}
	}

	if !vRegister.IsValid() || vRegister.IsNil() {
		return xerror.Fmt("[%#v, %#v] 没有找到匹配的interface", handler, vRegister.Interface())
	}

	vRegister.Call([]reflect.Value{reflect.ValueOf(server), reflect.ValueOf(handler)})
	return
}
