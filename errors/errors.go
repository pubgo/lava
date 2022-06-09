package errors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// MaxCode [0,1000]为系统错误, 业务错误code都大于1000
const MaxCode = 1000

// IsBisErr Business error
func IsBisErr(err error) bool {
	var err1 = FromError(err)
	if err1 == nil {
		return false
	}

	if err1.Reason == "" {
		return false
	}

	return true
}

// GRPCStatus 实现grpc status的GRPCStatus接口
func (x *Error) GRPCStatus() *status.Status {
	var dt, err = json.Marshal(x)
	xerror.Panic(err, "errors json marshal failed")
	return status.New(codes.Code(x.Code), string(dt))
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
	if err == nil {
		return false
	}

	if err1, ok := err.(*Error); ok {
		return x.Code == err1.Code && x.Reason == err1.Reason
	}

	if se := new(Error); errors.As(err, &se) {
		return se.Reason == x.Reason && se.Code == x.Code
	}

	return false
}

func (x *Error) As(target interface{}) bool {
	if target == nil {
		return false
	}

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
	return fmt.Sprintf("error: code=%d reason=%s message=%s metadata=%v", x.Code, x.Reason, x.Message, x.Metadata)
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
func New(reason string, msg string, args ...interface{}) *Error {
	return &Error{Reason: reason, Message: fmt.Sprintf(msg, args...), Metadata: make(map[string]string)}
}

// BadRequest generates a 400 error.
func BadRequest(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.InvalidArgument)
	return err
}

// Unauthorized generates a 401 error.
func Unauthorized(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Unauthenticated)
	return err
}

// Forbidden generates a 403 error.
func Forbidden(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.PermissionDenied)
	return err
}

// NotFound generates a 404 error.
func NotFound(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.NotFound)
	return err
}

// Timeout generates a 408 error.
func Timeout(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.DeadlineExceeded)
	return err
}

// Conflict generates a 409 error.
func Conflict(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.AlreadyExists)
	return err
}

// InternalServerError generates a 500 error.
func InternalServerError(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Internal)
	return err
}

// Cancelled The operation was cancelled, typically by the caller.
// HTTP Mapping: 499 Srv Closed Request
func Cancelled(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Canceled)
	return err
}

// Unknown error.
// HTTP Mapping: 500 Internal Grpc Error
func Unknown(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Unknown)
	return err
}

// InvalidArgument The client specified an invalid argument.
// HTTP Mapping: 400 Bad Request
func InvalidArgument(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.InvalidArgument)
	return err
}

// DeadlineExceeded The deadline expired before the operation could complete.
// HTTP Mapping: 504 Gateway Timeout
func DeadlineExceeded(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.DeadlineExceeded)
	return err
}

// AlreadyExists The entity that a client attempted to create (e.g., file or directory) already exists.
// HTTP Mapping: 409 Conflict
func AlreadyExists(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.AlreadyExists)
	return err
}

// PermissionDenied The caller does not have permission to execute the specified operation.
// HTTP Mapping: 403 Forbidden
func PermissionDenied(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.PermissionDenied)
	return err
}

// ResourceExhausted Some resource has been exhausted, perhaps a per-user quota, or
// perhaps the entire file system is out of space.
// HTTP Mapping: 429 Too Many Requests
func ResourceExhausted(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.ResourceExhausted)
	return err
}

// FailedPrecondition The operation was rejected because the system is not in a state
// required for the operation's execution.
// HTTP Mapping: 400 Bad Request
func FailedPrecondition(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.FailedPrecondition)
	return err
}

// Aborted The operation was aborted, typically due to a concurrency issue such as
// a sequencer check failure or transaction abort.
// HTTP Mapping: 409 Conflict
func Aborted(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Aborted)
	return err
}

// OutOfRange The operation was attempted past the valid range.  E.g., seeking or
// reading past end-of-file.
// HTTP Mapping: 400 Bad Request
func OutOfRange(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.OutOfRange)
	return err
}

// Unimplemented The operation is not implemented or is not supported/enabled in this service.
// HTTP Mapping: 501 Not Implemented
func Unimplemented(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Unimplemented)
	return err
}

// Internal This means that some invariants expected by the
// underlying system have been broken.  This error code is reserved
// for serious errors.
//
// HTTP Mapping: 500 Internal Grpc Error
func Internal(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Internal)
	return err
}

// Unavailable The service is currently unavailable.
// HTTP Mapping: 503 Service Unavailable
func Unavailable(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.Unavailable)
	return err
}

// DataLoss Unrecoverable data loss or corruption.
// HTTP Mapping: 500 Internal Grpc Error
func DataLoss(err *Error) error {
	if err == nil {
		return nil
	}

	err.Code = int32(codes.DataLoss)
	return err
}
