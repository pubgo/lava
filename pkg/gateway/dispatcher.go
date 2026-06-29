package gateway

import (
	"context"
	"io"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
) (header metadata.MD, trailer metadata.MD, err error) {
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

	return d.Dispatch(ctx, backend, frontend, op, in)
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
) (header metadata.MD, trailer metadata.MD, err error) {
	if op == nil {
		return nil, nil, errors.New("operation is nil")
	}
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
		err = d.dispatchBidi(ctx, backend, frontend, op)
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
	out := op.OutputType.New().Interface()
	var header metadata.MD
	var trailer metadata.MD
	if err := backend.Invoke(ctx, op.FullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer)); err != nil {
		return nil, nil, err
	}
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
			if hdr, headerErr := localStream.Header(); headerErr == nil {
				if sendErr := frontend.SendHeader(hdr); sendErr != nil {
					if !isDuplicateHeaderError(sendErr) {
						return errors.WrapCaller(sendErr)
					}
				}
			}
			headerSent = true
		}

		if err = frontend.SendMsg(out); err != nil {
			return errors.WrapCaller(err)
		}
	}

	if !headerSent {
		if hdr, headerErr := localStream.Header(); headerErr == nil {
			if sendErr := frontend.SendHeader(hdr); sendErr != nil {
				if !isDuplicateHeaderError(sendErr) {
					return errors.WrapCaller(sendErr)
				}
			}
		}
	}

	frontend.SetTrailer(localStream.Trailer())
	return nil
}

func (d *Dispatcher) dispatchClientStream(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
) error {
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

	if hdr, headerErr := localStream.Header(); headerErr == nil {
		if sendErr := frontend.SendHeader(hdr); sendErr != nil && !isDuplicateHeaderError(sendErr) {
			return errors.WrapCaller(sendErr)
		}
	}

	if err = frontend.SendMsg(out); err != nil {
		return errors.WrapCaller(err)
	}

	frontend.SetTrailer(localStream.Trailer())
	return nil
}

func (d *Dispatcher) dispatchBidi(
	ctx context.Context,
	backend Backend,
	frontend FrontendStream,
	op *Operation,
) error {
	clientCtx, clientCancel := context.WithCancel(ctx)
	defer clientCancel()

	localStream, err := backend.NewStream(clientCtx, op.StreamDesc, op.FullMethod)
	if err != nil {
		return errors.WrapCaller(err)
	}

	s2cErrChan := pumpFrontendToBackend(op.InputType, frontend, localStream)
	c2sErrChan := pumpBackendToFrontend(op.OutputType, localStream, frontend)

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
			frontend.SetTrailer(localStream.Trailer())
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
// stream to the frontend server stream. Unlike the proxy forwarder it never
// calls ClientStream.Header(): the inprocgrpc backend may deliver data frames
// without an explicit header frame, which would make Header() block waiting for
// a frame that never arrives. Frontends emit their own response headers on the
// first SendMsg, so backend header propagation is not required here.
func pumpBackendToFrontend(out protoreflect.MessageType, src grpc.ClientStream, dst FrontendStream) chan error {
	ret := make(chan error, 1)
	go func() {
		for {
			msg := out.New().Interface()
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
