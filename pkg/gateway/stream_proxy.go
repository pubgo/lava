// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package gateway

import (
	"context"
	"io"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/internal/logutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	clientStreamDescForProxying = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

func TransparentHandler(cli grpc.ClientConnInterface, inType, outType protoreflect.MessageType, opts ...grpc.CallOption) grpc.StreamHandler {
	streamer := &handler{cli: cli, opts: opts, inType: inType, outType: outType}
	return streamer.handler
}

type handler struct {
	cli             grpc.ClientConnInterface
	inType, outType protoreflect.MessageType
	opts            []grpc.CallOption
}

// handler is where the real magic of proxying happens.
// It is invoked like any gRPC server stream and uses the emptypb.Empty type server
// to proxy calls between the input and output streams.
func (s *handler) handler(srv interface{}, serverStream grpc.ServerStream) error {
	// little bit of gRPC internals never hurt anyone
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}

	clientCtx, clientCancel := context.WithCancel(serverStream.Context())
	defer clientCancel()
	clientStream, err := s.cli.NewStream(clientCtx, clientStreamDescForProxying, fullMethodName, s.opts...)
	if err != nil {
		return errors.WrapCaller(err)
	}

	// Explicitly *do not close* s2cErrChan and c2sErrChan, otherwise the select below will not terminate.
	// Channels do not have to be closed, it is just a control flow mechanism, see
	// https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ
	s2cErrChan := forwardServerToClient(s.inType, serverStream, clientStream)
	c2sErrChan := forwardClientToServer(s.outType, clientStream, serverStream)
	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				logutil.HandlerErr(clientStream.CloseSend())
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				clientCancel()
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}

func forwardClientToServer(out protoreflect.MessageType, src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := out.New().Interface()
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely.
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func forwardServerToClient(in protoreflect.MessageType, src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := in.New().Interface()
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}
