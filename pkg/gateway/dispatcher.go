package gateway

import (
	"context"
	"io"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

	s2cErrChan := forwardServerToClient(op.InputType, frontend, localStream)
	c2sErrChan := forwardClientToServer(op.OutputType, localStream, frontend)

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
