package gateway

import (
	"context"
	"io"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type dispatchConfig struct {
	// propagateBackendHeaders asks the backend→frontend pump to call
	// ClientStream.Header() before the first response message. Safe for remote
	// proxies; must stay false for inprocgrpc (Header can block forever).
	propagateBackendHeaders bool
}

// DispatchOption configures Dispatch / DispatchFrontend behavior.
type DispatchOption func(*dispatchConfig)

// WithPropagateBackendHeaders enables forwarding backend response headers on
// the first streamed message. Use for remote ClientConn backends only.
func WithPropagateBackendHeaders() DispatchOption {
	return func(c *dispatchConfig) { c.propagateBackendHeaders = true }
}

func applyDispatchOptions(opts []DispatchOption) dispatchConfig {
	var cfg dispatchConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// DispatchFrontend drives a frontend grpc.ServerStream end-to-end. For unary and
// server-streaming RPCs it pre-reads the request message via RecvMsg before
// dispatching; client-streaming and bidi read inside the pump. This is the
// shared entrypoint for frontends whose request surface is already a
// grpc.ServerStream (native gRPC passthrough, websocket, zrpc).
func (d *Dispatcher) DispatchFrontend(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
	opts ...DispatchOption,
) (header, trailer metadata.MD, err error) {
	if op == nil {
		return nil, nil, errors.New("operation is nil")
	}

	preReadRequest := op.StreamDesc == nil ||
		(op.StreamDesc.ServerStreams && !op.StreamDesc.ClientStreams)

	var in any
	if preReadRequest {
		req := op.InputType.New().Interface()
		if err = frontend.RecvMsg(req); err != nil {
			return nil, nil, err
		}
		in = req
	}

	return d.Dispatch(ctx, backend, frontend, op, in, opts...)
}

// Dispatch connects a frontend stream to the backend for the given operation.
// For unary and server-streaming RPCs, in must already be populated by the
// frontend (RecvMsg on the request message).
func (d *Dispatcher) Dispatch(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
	in any,
	opts ...DispatchOption,
) (header, trailer metadata.MD, err error) {
	if op == nil {
		return nil, nil, errors.New("operation is nil")
	}
	cfg := applyDispatchOptions(opts)
	if op.StreamDesc == nil {
		header, trailer, err = d.dispatchUnary(ctx, backend, frontend, op, in)
		return header, trailer, err
	}

	desc := op.StreamDesc
	switch {
	case desc.ServerStreams && !desc.ClientStreams:
		err = d.dispatchServerStream(ctx, backend, frontend, op, in)
		return nil, nil, err
	case desc.ClientStreams && !desc.ServerStreams:
		err = d.dispatchClientStream(ctx, backend, frontend, op)
		return nil, nil, err
	case desc.ClientStreams && desc.ServerStreams:
		err = d.dispatchBidi(ctx, backend, frontend, op, cfg)
		return nil, nil, err
	default:
		return nil, nil, errors.Errorf("unsupported stream mode: %s", op.FullMethod)
	}
}

func (d *Dispatcher) dispatchUnary(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
	in any,
) (metadata.MD, metadata.MD, error) {
	ctx, bag := withMDBag(ctx)
	out := op.OutputType.New().Interface()
	var header metadata.MD
	var trailer metadata.MD
	if err := backend.Invoke(ctx, op.FullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer)); err != nil {
		return nil, nil, err
	}
	header = mergeMD(header, bag.header)
	trailer = mergeMD(trailer, bag.trailer)
	if err := frontend.SendMsg(out); err != nil {
		return header, trailer, errors.WrapCaller(err)
	}
	return header, trailer, nil
}

func (d *Dispatcher) dispatchServerStream(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
	in any,
) error {
	ctx, bag := withMDBag(ctx)
	localStream, err := backend.NewStream(ctx, op.StreamDesc, op.FullMethod)
	if err != nil {
		return errors.WrapCaller(err)
	}

	if err = localStream.SendMsg(in); err != nil {
		return errors.WrapCaller(err)
	}
	if err = localStream.CloseSend(); err != nil {
		return errors.WrapCaller(err)
	}

	headerSent := false
	if streamHTTP, ok := frontend.(*streamHTTP); ok {
		streamHTTP.responseStream = true
	}

	for {
		out := op.OutputType.New().Interface()
		err = localStream.RecvMsg(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.WrapCaller(err)
		}

		if !headerSent {
			hdr, _ := localStream.Header()
			hdr = mergeMD(hdr, bag.header)
			if sendErr := frontend.SendHeader(hdr); sendErr != nil {
				if !isDuplicateHeaderError(sendErr) {
					return errors.WrapCaller(sendErr)
				}
			}
			headerSent = true
		}

		if err = frontend.SendMsg(out); err != nil {
			return errors.WrapCaller(err)
		}
	}

	if !headerSent {
		hdr, _ := localStream.Header()
		hdr = mergeMD(hdr, bag.header)
		if sendErr := frontend.SendHeader(hdr); sendErr != nil {
			if !isDuplicateHeaderError(sendErr) {
				return errors.WrapCaller(sendErr)
			}
		}
	}

	frontend.SetTrailer(mergeMD(localStream.Trailer(), bag.trailer))
	return nil
}

func (d *Dispatcher) dispatchClientStream(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
) error {
	ctx, bag := withMDBag(ctx)
	localStream, err := backend.NewStream(ctx, op.StreamDesc, op.FullMethod)
	if err != nil {
		return errors.WrapCaller(err)
	}

	in := op.InputType.New().Interface()
	for {
		if err = frontend.RecvMsg(in); err == io.EOF {
			break
		} else if err != nil {
			return errors.WrapCaller(err)
		}
		if err = localStream.SendMsg(in); err != nil {
			return errors.WrapCaller(err)
		}
		in = op.InputType.New().Interface()
	}

	if err = localStream.CloseSend(); err != nil {
		return errors.WrapCaller(err)
	}

	out := op.OutputType.New().Interface()
	if err = localStream.RecvMsg(out); err != nil {
		return errors.WrapCaller(err)
	}

	if hdr, headerErr := localStream.Header(); headerErr == nil || len(bag.header) > 0 {
		hdr = mergeMD(hdr, bag.header)
		if sendErr := frontend.SendHeader(hdr); sendErr != nil && !isDuplicateHeaderError(sendErr) {
			return errors.WrapCaller(sendErr)
		}
	}

	if err = frontend.SendMsg(out); err != nil {
		return errors.WrapCaller(err)
	}

	frontend.SetTrailer(mergeMD(localStream.Trailer(), bag.trailer))
	return nil
}

func (d *Dispatcher) dispatchBidi(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
	cfg dispatchConfig,
) error {
	ctx, bag := withMDBag(ctx)
	clientCtx, clientCancel := context.WithCancel(ctx)
	defer clientCancel()

	localStream, err := backend.NewStream(clientCtx, op.StreamDesc, op.FullMethod)
	if err != nil {
		return errors.WrapCaller(err)
	}

	s2cErrChan := pumpFrontendToBackend(op.InputType, frontend, localStream)
	c2sErrChan := pumpBackendToFrontend(op.OutputType, localStream, frontend, cfg.propagateBackendHeaders)

	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				if err = localStream.CloseSend(); err != nil {
					return errors.WrapCaller(err)
				}
			} else if s2cErr != nil {
				clientCancel()
				return errors.WrapCaller(s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			frontend.SetTrailer(mergeMD(localStream.Trailer(), bag.trailer))
			if c2sErr != io.EOF && c2sErr != nil {
				return c2sErr
			}
			return nil
		}
	}

	return errors.New("gRPC dispatch bidi should never reach this stage")
}

// pumpFrontendToBackend forwards request messages from the frontend server
// stream to the backend client stream until the frontend signals EOF.
func pumpFrontendToBackend(in protoreflect.MessageType, src FrontendStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		for {
			msg := in.New().Interface()
			if err := src.RecvMsg(msg); err != nil {
				ret <- err
				return
			}
			if err := dst.SendMsg(msg); err != nil {
				ret <- err
				return
			}
		}
	}()
	return ret
}

// pumpBackendToFrontend forwards response messages from the backend client
// stream to the frontend server stream.
//
// When propagateHeader is false (default for Mux/inproc), Header() is never
// called: inprocgrpc may deliver data without an explicit header frame, and
// Header() would block. Frontends emit their own headers on first SendMsg.
//
// When propagateHeader is true (remote TransparentHandler), the first response
// message triggers Header()+SendHeader before SendMsg.
func pumpBackendToFrontend(out protoreflect.MessageType, src grpc.ClientStream, dst FrontendStream, propagateHeader bool) chan error {
	ret := make(chan error, 1)
	go func() {
		for i := 0; ; i++ {
			msg := out.New().Interface()
			if err := src.RecvMsg(msg); err != nil {
				ret <- err
				return
			}
			if propagateHeader && i == 0 {
				if md, err := src.Header(); err == nil {
					if sendErr := dst.SendHeader(md); sendErr != nil && !isDuplicateHeaderError(sendErr) {
						ret <- sendErr
						return
					}
				} else {
					ret <- err
					return
				}
			}
			if err := dst.SendMsg(msg); err != nil {
				ret <- err
				return
			}
		}
	}()
	return ret
}
