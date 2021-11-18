package errors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/x/jsonx"
)

//func (e *Error) Error() string {
//	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v", e.Code, e.Reason, e.Message, e.Metadata)
//}

// GRPCStatus returns the Status represented by se.
func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(codes.Code(e.Code), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Reason,
			Metadata: e.Metadata,
		})
	return s
}

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Reason == e.Reason
	}
	return false
}

// WithDetail with an MD formed by the mapping of key, value.
func (e *Error) WithMetadata(m map[string]string) *Error {
	err := proto.Clone(e).(*Error)
	err.Metadata = m
	return err
}

func (x *Error) Error() string {
	var dt, err = jsonx.Marshal(x)
	if err != nil {
		return err.Error()
	}
	return string(dt)
}

func (x *Error) As(target interface{}) bool {
	t1 := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if err, ok := t1.(*Error); ok {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
		return true
	}
	return false
}

// Is matches each error in the chain with the target value.
func (x *Error) Is1(target error) bool {
	err, ok := target.(*Error)
	if ok {
		return x.Code == err.Code
	}
	return false
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

// Code returns the status code.
func Code1(err error) int32 {
	if err == nil {
		return 0 // ok
	}

	if se := new(Error); errors.As(err, &se) {
		return se.Code
	}
	return 2 // unknown
}

//// Parse tries to parse a JSON string into an error. If that
//// fails, it will set the given string as the error Status.
//func Parse(err string) *Error {
//	e := new(Error)
//	err1 := json.Unmarshal(strutil.ToBytes(err), e)
//	if err1 != nil {
//		e.Status = err
//	}
//	return e
//}
//

// New generates a custom error.
func New(reason string, code int32, msg string, args ...interface{}) *Error {
	return &Error{
		Reason:  reason,
		Code:    code,
		Message: fmt.Sprintf(msg, args...),
	}
}

//
//// BadRequest generates a 400 error.
//func BadRequest(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   400,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Unauthorized generates a 401 error.
//func Unauthorized(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   401,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Forbidden generates a 403 error.
//func Forbidden(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   403,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// NotFound generates a 404 error.
//func NotFound(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   404,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// MethodNotAllowed generates a 405 error.
//func MethodNotAllowed(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   405,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Timeout generates a 408 error.
//func Timeout(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   408,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Conflict generates a 409 error.
//func Conflict(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   409,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// InternalServerError generates a 500 error.
//func InternalServerError(id, format string, a ...interface{}) error {
//	return &Error{
//		Id:     id,
//		Code:   500,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//type IgnorableError struct {
//	Err       string `json:"error"`
//	Ignorable bool   `json:"ignorable"`
//}
//
//func (e *IgnorableError) Error() string {
//	b, _ := json.Marshal(e)
//	return string(b)
//}
//
//// UnwrapIgnorableError tries to parse a JSON string into an error. If that
//// fails, it will set the given string as the error Status.
//func UnwrapIgnorableError(err string) (bool, string) {
//	if err == "" {
//		return false, err
//	}
//
//	igErr := new(IgnorableError)
//	uErr := json.Unmarshal([]byte(err), igErr)
//	if uErr != nil {
//		return false, err
//	}
//
//	return igErr.Ignorable, igErr.Err
//}
//
//// IgnoreError generates a -1 error.
//func WrapIgnorableError(err error) error {
//	if err == nil {
//		return nil
//	}
//
//	return &IgnorableError{
//		Err:       err.Error(),
//		Ignorable: true,
//	}
//}
//
//// Cancelled The operation was cancelled, typically by the caller.
//// HTTP Mapping: 499 Client Closed Request
//func Cancelled(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   1,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//

// Unknown error.
// HTTP Mapping: 500 Internal Grpc Error
func Unknown(reason, format string, a ...interface{}) error {
	return &Error{
		Code:    2,
		Reason:  reason,
		Message: fmt.Sprintf(format, a...),
	}
}

//
//// InvalidArgument The client specified an invalid argument.
//// HTTP Mapping: 400 Bad Request
//func InvalidArgument(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   3,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// DeadlineExceeded The deadline expired before the operation could complete.
//// HTTP Mapping: 504 Gateway Timeout
//func DeadlineExceeded(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   4,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// AlreadyExists The entity that a client attempted to create (e.g., file or directory) already exists.
//// HTTP Mapping: 409 Conflict
//func AlreadyExists(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   6,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// PermissionDenied The caller does not have permission to execute the specified operation.
//// HTTP Mapping: 403 Forbidden
//func PermissionDenied(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   7,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// ResourceExhausted Some resource has been exhausted, perhaps a per-user quota, or
//// perhaps the entire file system is out of space.
//// HTTP Mapping: 429 Too Many Requests
//func ResourceExhausted(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   8,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// FailedPrecondition The operation was rejected because the system is not in a state
//// required for the operation's execution.
//// HTTP Mapping: 400 Bad Request
//func FailedPrecondition(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   9,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Aborted The operation was aborted, typically due to a concurrency issue such as
//// a sequencer check failure or transaction abort.
//// HTTP Mapping: 409 Conflict
//func Aborted(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   10,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// OutOfRange The operation was attempted past the valid range.  E.g., seeking or
//// reading past end-of-file.
//// HTTP Mapping: 400 Bad Request
//func OutOfRange(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   11,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Unimplemented The operation is not implemented or is not supported/enabled in this service.
//// HTTP Mapping: 501 Not Implemented
//func Unimplemented(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   12,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Internal This means that some invariants expected by the
//// underlying system have been broken.  This error code is reserved
//// for serious errors.
////
//// HTTP Mapping: 500 Internal Grpc Error
//func Internal(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   13,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// Unavailable The service is currently unavailable.
//// HTTP Mapping: 503 Service Unavailable
//func Unavailable(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   14,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
//
//// DataLoss Unrecoverable data loss or corruption.
//// HTTP Mapping: 500 Internal Grpc Error
//func DataLoss(id, format string, a ...interface{}) error {
//	return &Error{
//		Code:   15,
//		Id:     id,
//		Status: fmt.Sprintf(format, a...),
//	}
//}
