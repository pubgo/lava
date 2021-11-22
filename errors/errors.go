package errors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/pubgo/xerror"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// GRPCStatus 实现grpc status的GRPCStatus接口
func (x *Error) GRPCStatus() *status.Status {
	s, err := status.New(codes.Code(x.Code), x.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   x.Reason,
			Metadata: x.Metadata,
		})
	xerror.Panic(err)
	return s
}

// HTTPStatus returns the Status represented by se.
func (x *Error) HTTPStatus() int {
	switch x.Code {
	case 0:
		return http.StatusOK
	case 1:
		return http.StatusInternalServerError
	case 2:
		return http.StatusInternalServerError
	case 3:
		return http.StatusBadRequest
	case 4:
		return http.StatusRequestTimeout
	case 5:
		return http.StatusNotFound
	case 6:
		return http.StatusConflict
	case 7:
		return http.StatusForbidden
	case 8:
		return http.StatusTooManyRequests
	case 9:
		return http.StatusPreconditionFailed
	case 10:
		return http.StatusConflict
	case 11:
		return http.StatusBadRequest
	case 12:
		return http.StatusNotImplemented
	case 13:
		return http.StatusInternalServerError
	case 14:
		return http.StatusServiceUnavailable
	case 15:
		return http.StatusInternalServerError
	case 16:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Is matches each error in the chain with the target value.
func (x *Error) Is(err error) bool {
	if err1, ok := err.(*Error); ok {
		return x.Code == err1.Code && x.Reason == err1.Reason
	}

	if se := new(Error); errors.As(err, &se) {
		return se.Reason == x.Reason && se.Code == x.Code
	}
	return false
}

func (x *Error) As(target interface{}) bool {
	t1 := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if err, ok := t1.(*Error); ok {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
		return true
	}
	return false
}

// WithMetadata with an MD formed by the mapping of key, value.
func (x *Error) WithMetadata(m map[string]string) *Error {
	err := proto.Clone(x).(*Error)
	err.Metadata = m
	return err
}

func (x *Error) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v", x.Code, x.Reason, x.Message, x.Metadata)
}

// Code returns the status code.
func Code(err error) int32 {
	if err == nil {
		return 0 // ok
	}

	if se := new(Error); errors.As(err, &se) {
		return se.Code
	}
	return 2 // unknown
}

// New generates a custom error.
func New(reason string, code int32, msg string, args ...interface{}) *Error {
	return &Error{
		Reason:  reason,
		Code:    code,
		Message: fmt.Sprintf(msg, args...),
	}
}

// BadRequest generates a 400 error.
func BadRequest(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.InvalidArgument),
		Message: fmt.Sprintf(format, a...),
	}
}

// Unauthorized generates a 401 error.
func Unauthorized(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.Unauthenticated),
		Message: fmt.Sprintf(format, a...),
	}
}

// Forbidden generates a 403 error.
func Forbidden(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.PermissionDenied),
		Message: fmt.Sprintf(format, a...),
	}
}

// NotFound generates a 404 error.
func NotFound(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.NotFound),
		Message: fmt.Sprintf(format, a...),
	}
}

// Timeout generates a 408 error.
func Timeout(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.DeadlineExceeded),
		Message: fmt.Sprintf(format, a...),
	}
}

// Conflict generates a 409 error.
func Conflict(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.AlreadyExists),
		Message: fmt.Sprintf(format, a...),
	}
}

// InternalServerError generates a 500 error.
func InternalServerError(id, format string, a ...interface{}) error {
	return &Error{
		Reason:  id,
		Code:    int32(codes.Internal),
		Message: fmt.Sprintf(format, a...),
	}
}

// Cancelled The operation was cancelled, typically by the caller.
// HTTP Mapping: 499 Client Closed Request
func Cancelled(id, format string, a ...interface{}) error {
	return &Error{
		Code:    1,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// Unknown error.
// HTTP Mapping: 500 Internal Grpc Error
func Unknown(reason, format string, a ...interface{}) error {
	return &Error{
		Code:    2,
		Reason:  reason,
		Message: fmt.Sprintf(format, a...),
	}
}

// InvalidArgument The client specified an invalid argument.
// HTTP Mapping: 400 Bad Request
func InvalidArgument(id, format string, a ...interface{}) error {
	return &Error{
		Code:    3,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// DeadlineExceeded The deadline expired before the operation could complete.
// HTTP Mapping: 504 Gateway Timeout
func DeadlineExceeded(id, format string, a ...interface{}) error {
	return &Error{
		Code:    4,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// AlreadyExists The entity that a client attempted to create (e.g., file or directory) already exists.
// HTTP Mapping: 409 Conflict
func AlreadyExists(id, format string, a ...interface{}) error {
	return &Error{
		Code:    6,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// PermissionDenied The caller does not have permission to execute the specified operation.
// HTTP Mapping: 403 Forbidden
func PermissionDenied(id, format string, a ...interface{}) error {
	return &Error{
		Code:    7,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// ResourceExhausted Some resource has been exhausted, perhaps a per-user quota, or
// perhaps the entire file system is out of space.
// HTTP Mapping: 429 Too Many Requests
func ResourceExhausted(id, format string, a ...interface{}) error {
	return &Error{
		Code:    8,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// FailedPrecondition The operation was rejected because the system is not in a state
// required for the operation's execution.
// HTTP Mapping: 400 Bad Request
func FailedPrecondition(id, format string, a ...interface{}) error {
	return &Error{
		Code:    9,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// Aborted The operation was aborted, typically due to a concurrency issue such as
// a sequencer check failure or transaction abort.
// HTTP Mapping: 409 Conflict
func Aborted(id, format string, a ...interface{}) error {
	return &Error{
		Code:    10,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// OutOfRange The operation was attempted past the valid range.  E.g., seeking or
// reading past end-of-file.
// HTTP Mapping: 400 Bad Request
func OutOfRange(id, format string, a ...interface{}) error {
	return &Error{
		Code:    11,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// Unimplemented The operation is not implemented or is not supported/enabled in this service.
// HTTP Mapping: 501 Not Implemented
func Unimplemented(id, format string, a ...interface{}) error {
	return &Error{
		Code:    12,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// Internal This means that some invariants expected by the
// underlying system have been broken.  This error code is reserved
// for serious errors.
//
// HTTP Mapping: 500 Internal Grpc Error
func Internal(id, format string, a ...interface{}) error {
	return &Error{
		Code:    13,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// Unavailable The service is currently unavailable.
// HTTP Mapping: 503 Service Unavailable
func Unavailable(id, format string, a ...interface{}) error {
	return &Error{
		Code:    14,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}

// DataLoss Unrecoverable data loss or corruption.
// HTTP Mapping: 500 Internal Grpc Error
func DataLoss(id, format string, a ...interface{}) error {
	return &Error{
		Code:    15,
		Reason:  id,
		Message: fmt.Sprintf(format, a...),
	}
}
