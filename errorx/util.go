package errorx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
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

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *errorpb.Error {
	if err == nil {
		return nil
	}

	var ce errors.ErrCode
	if errors.As(err, &ce) {
		return &errorpb.Error{
			Service:   version.Project(),
			Version:   version.Version(),
			Code:      ce.Code(),
			BizCode:   ce.BizCode(),
			Reason:    ce.Reason(),
			ErrMsg:    err.Error(),
			ErrDetail: fmt.Sprintf("%#v", err),
			Tags:      ce.Tags(),
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
			Code:      errorpb.Code(gs.GRPCStatus().Code())}
	}

	return &errorpb.Error{
		ErrMsg:    err.Error(),
		ErrDetail: fmt.Sprintf("%#v", err),
		Reason:    err.Error(),
		Code:      errorpb.Code_Unknown,
	}
}

// Convert 内部转换，为了让err=nil的时候，监控数据里有OK信息
func Convert(err *errorpb.Error) *status.Status {
	if generic.IsNil(err) {
		return status.New(codes.OK, "OK")
	}

	var st, err1 = status.New(codes.Code(err.Code), err.ErrMsg).WithDetails(err)
	if err1 != nil {
		log.Err(err1).Any("convert-err", err).Msg("failed to convert error detail")
		status.New(codes.Internal, err1.Error())
	}
	return st
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

func lavaError(code int32) codes.Code {
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
