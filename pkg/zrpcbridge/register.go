package zrpcbridge

import (
	"context"
	"strings"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/pkg/zrpc"
)

// Config configures NATS/zrpc bindings to a gateway.Mux backend.
type Config struct {
	// Queue is the NATS queue group name for all registered methods.
	Queue string
	// SubjectPrefix is prepended to each gRPC full method (without leading slash).
	// Default: "svc." — e.g. "/pkg.v1.Service/Method" → "svc.pkg.v1.Service/Method".
	SubjectPrefix string
}

// Subject returns the NATS subject for a gRPC full method name.
func Subject(fullMethod, prefix string) string {
	if prefix == "" {
		prefix = "svc."
	}
	return prefix + strings.TrimPrefix(fullMethod, "/")
}

// RegisterMux binds all methods registered on mux to zrpc.Server NATS subscriptions.
// Business handlers stay on gateway.Mux; zrpc is a separate transport bridge.
func RegisterMux(srv *zrpc.Server, mux *gateway.Mux, cfg Config) error {
	if srv == nil {
		return errors.New("zrpc server is nil")
	}
	if mux == nil {
		return errors.New("gateway mux is nil")
	}
	if cfg.Queue == "" {
		return errors.New("zrpc queue is required")
	}

	prefix := cfg.SubjectPrefix
	if prefix == "" {
		prefix = "svc."
	}

	for _, route := range mux.Routes() {
		op := route.Operation
		if op == nil {
			continue
		}
		subject := Subject(route.FullMethod, prefix)
		if op.StreamDesc == nil {
			if err := registerUnary(srv, mux, subject, cfg.Queue, route.FullMethod, op); err != nil {
				return err
			}
			continue
		}
		if err := registerStream(srv, mux, subject, cfg.Queue, route.FullMethod, op); err != nil {
			return err
		}
	}
	return nil
}

func registerUnary(srv *zrpc.Server, mux *gateway.Mux, subject, queue, fullMethod string, op *gateway.Operation) error {
	return zrpc.RegisterUnary[proto.Message, proto.Message](
		srv,
		subject,
		queue,
		func() proto.Message { return op.InputType.New().Interface() },
		func(ctx context.Context, req proto.Message) (proto.Message, error) {
			frontend := &streamUnary{ctx: ctx, fullMethod: fullMethod}
			_, _, err := mux.Dispatch(frontend.Context(), frontend, op, req)
			if err != nil {
				return nil, err
			}
			return frontend.response, nil
		},
	)
}

func registerStream(srv *zrpc.Server, mux *gateway.Mux, subject, queue, fullMethod string, op *gateway.Operation) error {
	return zrpc.RegisterStream(srv, subject, queue, func(ctx context.Context, zstream *zrpc.ServerStream) error {
		stream := &streamAdapter{zstream: zstream, ctx: ctx, fullMethod: fullMethod}
		_, _, err := mux.DispatchFrontend(stream.Context(), stream, op)
		return err
	})
}
