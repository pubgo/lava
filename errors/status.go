package errors

import (
	errorV1 "github.com/pubgo/lava/gen/proto/errors/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

func (e *Error) StatusCancelled() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Canceled)
	return &Error{err: ee}
}

func (e *Error) StatusUnknown() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Unknown)
	return &Error{err: ee}
}

func (e *Error) StatusInvalidArgument() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.InvalidArgument)
	return &Error{err: ee}
}

func (e *Error) StatusBadRequest() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.InvalidArgument)
	return &Error{err: ee}
}

func (e *Error) StatusDeadlineExceeded() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.DeadlineExceeded)
	return &Error{err: ee}
}

func (e *Error) StatusTimeout() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.DeadlineExceeded)
	return &Error{err: ee}
}

func (e *Error) StatusNotFound() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.NotFound)
	return &Error{err: ee}
}

func (e *Error) StatusAlreadyExists() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.AlreadyExists)
	return &Error{err: ee}
}

func (e *Error) StatusConflict() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.AlreadyExists)
	return &Error{err: ee}
}

func (e *Error) StatusForbidden() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.PermissionDenied)
	return &Error{err: ee}
}

func (e *Error) StatusPermissionDenied() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.PermissionDenied)
	return &Error{err: ee}
}

func (e *Error) StatusResourceExhausted() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.ResourceExhausted)
	return &Error{err: ee}
}

func (e *Error) StatusFailedPrecondition() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.FailedPrecondition)
	return &Error{err: ee}
}

func (e *Error) StatusAborted() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Aborted)
	return &Error{err: ee}
}

func (e *Error) StatusOutOfRange() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.OutOfRange)
	return &Error{err: ee}
}

func (e *Error) StatusUnimplemented() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Unimplemented)
	return &Error{err: ee}
}

func (e *Error) StatusInternal() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Internal)
	return &Error{err: ee}
}

func (e *Error) StatusUnavailable() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Unavailable)
	return &Error{err: ee}
}

func (e *Error) StatusDataLoss() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.DataLoss)
	return &Error{err: ee}
}

func (e *Error) StatusUnauthorized() error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(codes.Unauthenticated)
	return &Error{err: ee}
}
