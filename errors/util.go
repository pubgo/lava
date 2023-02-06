package errors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Err2GrpcCode
// converts a standard Go error into its canonical code. Note that
// this is only used to translate the error returned by the server applications.
func Err2GrpcCode(err error) codes.Code {
	switch err {
	case nil:
		return codes.OK
	case io.EOF:
		return codes.OutOfRange
	case io.ErrClosedPipe, io.ErrNoProgress, io.ErrShortBuffer, io.ErrShortWrite, io.ErrUnexpectedEOF:
		return codes.FailedPrecondition
	case os.ErrInvalid:
		return codes.InvalidArgument
	case context.Canceled:
		return codes.Canceled
	case context.DeadlineExceeded:
		return codes.DeadlineExceeded
	}

	switch {
	case os.IsExist(err):
		return codes.AlreadyExists
	case os.IsNotExist(err):
		return codes.NotFound
	case os.IsPermission(err):
		return codes.PermissionDenied
	}
	return codes.Unknown
}

func Http2GrpcCode(code int32) codes.Code {
	switch code {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusPreconditionFailed:
		return codes.FailedPrecondition
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	}

	return codes.Unknown
}

func IsGrpcAcceptable(err error) bool {
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Internal, codes.Unavailable, codes.DataLoss:
		return false
	default:
		return true
	}
}

func IsMemoryErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "invalid memory address or nil pointer dereference")
}

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *errorpb.Error {
	switch err.(type) {
	case nil:
		return nil
	case *Error:
		return err.(*Error).Proto()
	case Error:
		return err.(Error).Proto()
	}

	var ce errors.ErrCode
	if errors.As(err, &ce) {
		return &errorpb.Error{
			Service:   version.Project(),
			Version:   version.Version(),
			Code:      uint32(ce.Code()),
			BizCode:   ce.BizCode(),
			Reason:    ce.Reason(),
			ErrMsg:    err.Error(),
			ErrDetail: fmt.Sprintf("%#v", err),
		}
	}

	// grpc error
	gs, ok := err.(interface{ GRPCStatus() *status.Status })
	if ok {
		if gs.GRPCStatus().Code() == codes.OK {
			return nil
		}

		details := gs.GRPCStatus().Details()
		if len(details) > 0 && details[0] != nil {
			if e, ok := details[0].(*errorpb.Error); ok && e != nil {
				return e
			}
		}

		return &errorpb.Error{
			ErrMsg:    err.Error(),
			ErrDetail: fmt.Sprintf("%v", gs.GRPCStatus().Details()),
			Reason:    gs.GRPCStatus().Message(),
			Code:      uint32(gs.GRPCStatus().Code())}
	}

	return &errorpb.Error{
		ErrMsg:    err.Error(),
		ErrDetail: fmt.Sprintf("%#v", err),
		Reason:    err.Error(),
		Code:      uint32(errorpb.Code_Unknown),
	}
}

// Convert 内部转换，为了让err=nil的时候，监控数据里有OK信息
func Convert(err error) *status.Status {
	if generic.IsNil(err) {
		return status.New(codes.OK, "OK")
	}

	if se, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
		return se.GRPCStatus()
	}

	switch err {
	case context.DeadlineExceeded:
		return status.New(codes.DeadlineExceeded, err.Error())
	case context.Canceled:
		return status.New(codes.Canceled, err.Error())
	}

	return status.New(codes.Unknown, err.Error())
}

// GrpcCodeToHTTP gRPC转HTTP Code
func GrpcCodeToHTTP(statusCode codes.Code) int {
	switch statusCode {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusServiceUnavailable
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func lavaError(err *Error) codes.Code {
	switch err {
	case nil:
		return codes.OK
	}

	switch err.code {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusPreconditionFailed:
		return codes.FailedPrecondition
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	}

	return codes.Unknown
}
