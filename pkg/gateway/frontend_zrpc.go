package gateway

import (
	"context"
	"strings"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/pkg/zrpc"
)

// ZrpcConfig configures NATS/zrpc frontend bindings for registered Mux methods.
type ZrpcConfig struct {
	// Queue is the NATS queue group name for all registered methods.
	Queue string
	// SubjectPrefix is prepended to each gRPC full method (without leading slash).
	// Default: "svc." — e.g. "/pkg.v1.Service/Method" → "svc.pkg.v1.Service/Method".
	SubjectPrefix string
}

// ZrpcSubject returns the NATS subject for a gRPC full method name.
func ZrpcSubject(fullMethod, prefix string) string {
	if prefix == "" {
		prefix = "svc."
	}
	return prefix + strings.TrimPrefix(fullMethod, "/")
}

// RegisterZrpc binds all methods registered on this Mux to a zrpc.Server.
// Handlers are dispatched through the unified Dispatcher, sharing the same
// backend as HTTP, WebSocket, and native gRPC frontends.
func (m *Mux) RegisterZrpc(srv *zrpc.Server, cfg ZrpcConfig) error {
	if srv == nil {
		return errors.New("zrpc server is nil")
	}
	if cfg.Queue == "" {
		return errors.New("zrpc queue is required")
	}

	prefix := cfg.SubjectPrefix
	if prefix == "" {
		prefix = "svc."
	}

	for _, mth := range m.opts.handlers {
		op := operationFromMethod(mth)
		if op == nil {
			continue
		}

		subject := ZrpcSubject(mth.grpcFullMethod, prefix)
		if op.StreamDesc == nil {
			if err := registerZrpcUnary(m, srv, subject, cfg.Queue, mth, op); err != nil {
				return err
			}
			continue
		}

		if err := registerZrpcStream(m, srv, subject, cfg.Queue, mth, op); err != nil {
			return err
		}
	}

	return nil
}

func registerZrpcUnary(m *Mux, srv *zrpc.Server, subject, queue string, mth *methodWrapper, op *Operation) error {
	return zrpc.RegisterUnary[proto.Message, proto.Message](
		srv,
		subject,
		queue,
		func() proto.Message { return mth.inputType.New().Interface().(proto.Message) },
		func(ctx context.Context, req proto.Message) (proto.Message, error) {
			frontend := &streamZrpcUnary{ctx: ctx, method: mth}
			_, _, err := m.dispatcher.Dispatch(frontend.Context(), m, frontend, op, req)
			if err != nil {
				return nil, err
			}
			return frontend.response, nil
		},
	)
}

func registerZrpcStream(m *Mux, srv *zrpc.Server, subject, queue string, mth *methodWrapper, op *Operation) error {
	return zrpc.RegisterStream(srv, subject, queue, func(ctx context.Context, zstream *zrpc.ServerStream) error {
		stream := &streamZrpc{zstream: zstream, ctx: ctx, method: mth}
		return dispatchZrpcStream(m, stream, op)
	})
}

func dispatchZrpcStream(m *Mux, stream *streamZrpc, op *Operation) error {
	preReadRequest := op.StreamDesc == nil ||
		(op.StreamDesc.ServerStreams && !op.StreamDesc.ClientStreams)

	var in any
	if preReadRequest {
		req := op.InputType.New().Interface()
		if err := stream.RecvMsg(req); err != nil {
			return err
		}
		in = req
	}

	_, _, err := m.dispatcher.Dispatch(stream.Context(), m, stream, op, in)
	return err
}
