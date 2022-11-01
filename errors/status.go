package errors

import (
	"google.golang.org/grpc/codes"
)

func (e Error) StatusCancelled() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Canceled)
	return e
}

func (e Error) StatusUnknown() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Unknown)
	return e
}

func (e Error) StatusInvalidArgument() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.InvalidArgument)
	return e
}

func (e Error) StatusBadRequest() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.InvalidArgument)
	return e
}

func (e Error) StatusDeadlineExceeded() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.DeadlineExceeded)
	return e
}

func (e Error) StatusTimeout() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.DeadlineExceeded)
	return e
}

func (e Error) StatusNotFound() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.NotFound)
	return e
}

func (e Error) StatusAlreadyExists() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.AlreadyExists)
	return e
}

func (e Error) StatusConflict() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.AlreadyExists)
	return e
}

func (e Error) StatusForbidden() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.PermissionDenied)
	return e
}

func (e Error) StatusPermissionDenied() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.PermissionDenied)
	return e
}

func (e Error) StatusResourceExhausted() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.ResourceExhausted)
	return e
}

func (e Error) StatusFailedPrecondition() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.FailedPrecondition)
	return e
}

func (e Error) StatusAborted() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Aborted)
	return e
}

func (e Error) StatusOutOfRange() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.OutOfRange)
	return e
}

func (e Error) StatusUnimplemented() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Unimplemented)
	return e
}

func (e Error) StatusInternal() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Internal)
	return e
}

func (e Error) StatusUnavailable() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Unavailable)
	return e
}

func (e Error) StatusDataLoss() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.DataLoss)
	return e
}

func (e Error) StatusUnauthorized() error {
	if e.err == nil {
		return nil
	}

	e.code = uint32(codes.Unauthenticated)
	return e
}
