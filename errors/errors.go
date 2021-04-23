package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pubgo/x/jsonx"
)

func (t *Error) Error() string {
	var dt, err = jsonx.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(dt)
}

// Is matches each error in the chain with the target value.
func (t *Error) Is(target error) bool {
	err, ok := target.(*Error)
	if ok {
		return t.Code == err.Code
	}
	return false
}

// HTTPStatus returns the Status represented by se.
func (t *Error) HTTPStatus() int {
	switch t.Code {
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
func Code(err error) int32 {
	if err == nil {
		return 0 // ok
	}

	if se := new(Error); errors.As(err, &se) {
		return se.Code
	}
	return 2 // unknown
}

// FromError returns status error.
func FromError(err error) (*Error, bool) {
	if se := new(Error); errors.As(err, &se) {
		return se, true
	}
	return nil, false
}

// New generates a custom error.
func New(id, Message string, code int32) *Error {
	return &Error{
		Id:      id,
		Code:    code,
		Message: Message,
	}
}

// Parse tries to parse a JSON string into an error. If that
// fails, it will set the given string as the error Message.
func Parse(err string) *Error {
	e := new(Error)
	err1 := json.Unmarshal([]byte(err), e)
	if err1 != nil {
		e.Message = err
	}
	return e
}

// BadRequest generates a 400 error.
func BadRequest(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    400,
		Message: fmt.Sprintf(format, a...),
	}
}

// Unauthorized generates a 401 error.
func Unauthorized(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    401,
		Message: fmt.Sprintf(format, a...),
	}
}

// Forbidden generates a 403 error.
func Forbidden(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    403,
		Message: fmt.Sprintf(format, a...),
	}
}

// NotFound generates a 404 error.
func NotFound(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    404,
		Message: fmt.Sprintf(format, a...),
	}
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    405,
		Message: fmt.Sprintf(format, a...),
	}
}

// Timeout generates a 408 error.
func Timeout(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    408,
		Message: fmt.Sprintf(format, a...),
	}
}

// Conflict generates a 409 error.
func Conflict(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    409,
		Message: fmt.Sprintf(format, a...),
	}
}

// InternalServerError generates a 500 error.
func InternalServerError(id, format string, a ...interface{}) error {
	return &Error{
		Id:      id,
		Code:    500,
		Message: fmt.Sprintf(format, a...),
	}
}

type IgnorableError struct {
	Err       string `json:"error"`
	Ignorable bool   `json:"ignorable"`
}

func (e *IgnorableError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// UnwrapIgnorableError tries to parse a JSON string into an error. If that
// fails, it will set the given string as the error Message.
func UnwrapIgnorableError(err string) (bool, string) {
	if err == "" {
		return false, err
	}

	igerr := new(IgnorableError)
	uerr := json.Unmarshal([]byte(err), igerr)
	if uerr != nil {
		return false, err
	}

	return igerr.Ignorable, igerr.Err
}

// IgnoreError generates a -1 error.
func WrapIgnorableError(err error) error {
	if err == nil {
		return nil
	}

	return &IgnorableError{
		Err:       err.Error(),
		Ignorable: true,
	}
}

// Cancelled The operation was cancelled, typically by the caller.
// HTTP Mapping: 499 Client Closed Request
func Cancelled(id, format string, a ...interface{}) error {
	return &Error{
		Code:    1,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsCancelled determines if err is an error which indicates a cancelled error.
// It supports wrapped errors.
func IsCancelled(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 1
	}
	return false
}

// Unknown error.
// HTTP Mapping: 500 Internal Server Error
func Unknown(id, format string, a ...interface{}) error {
	return &Error{
		Code:    2,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsUnknown determines if err is an error which indicates a unknown error.
// It supports wrapped errors.
func IsUnknown(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 2
	}
	return false
}

// InvalidArgument The client specified an invalid argument.
// HTTP Mapping: 400 Bad Request
func InvalidArgument(id, format string, a ...interface{}) error {
	return &Error{
		Code:    3,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsInvalidArgument determines if err is an error which indicates an invalid argument error.
// It supports wrapped errors.
func IsInvalidArgument(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 3
	}
	return false
}

// DeadlineExceeded The deadline expired before the operation could complete.
// HTTP Mapping: 504 Gateway Timeout
func DeadlineExceeded(id, format string, a ...interface{}) error {
	return &Error{
		Code:    4,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsDeadlineExceeded determines if err is an error which indicates a deadline exceeded error.
// It supports wrapped errors.
func IsDeadlineExceeded(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 4
	}
	return false
}

// IsNotFound determines if err is an error which indicates a not found error.
// It supports wrapped errors.
func IsNotFound(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 5
	}
	return false
}

// AlreadyExists The entity that a client attempted to create (e.g., file or directory) already exists.
// HTTP Mapping: 409 Conflict
func AlreadyExists(id, format string, a ...interface{}) error {
	return &Error{
		Code:    6,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsAlreadyExists determines if err is an error which indicates a already exsits error.
// It supports wrapped errors.
func IsAlreadyExists(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 6
	}
	return false
}

// PermissionDenied The caller does not have permission to execute the specified operation.
// HTTP Mapping: 403 Forbidden
func PermissionDenied(id, format string, a ...interface{}) error {
	return &Error{
		Code:    7,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsPermissionDenied determines if err is an error which indicates a permission denied error.
// It supports wrapped errors.
func IsPermissionDenied(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 7
	}
	return false
}

// ResourceExhausted Some resource has been exhausted, perhaps a per-user quota, or
// perhaps the entire file system is out of space.
// HTTP Mapping: 429 Too Many Requests
func ResourceExhausted(id, format string, a ...interface{}) error {
	return &Error{
		Code:    8,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsResourceExhausted determines if err is an error which indicates a resource exhausted error.
// It supports wrapped errors.
func IsResourceExhausted(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 8
	}
	return false
}

// FailedPrecondition The operation was rejected because the system is not in a state
// required for the operation's execution.
// HTTP Mapping: 400 Bad Request
func FailedPrecondition(id, format string, a ...interface{}) error {
	return &Error{
		Code:    9,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsFailedPrecondition determines if err is an error which indicates a failed precondition error.
// It supports wrapped errors.
func IsFailedPrecondition(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 9
	}
	return false
}

// Aborted The operation was aborted, typically due to a concurrency issue such as
// a sequencer check failure or transaction abort.
// HTTP Mapping: 409 Conflict
func Aborted(id, format string, a ...interface{}) error {
	return &Error{
		Code:    10,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsAborted determines if err is an error which indicates an aborted error.
// It supports wrapped errors.
func IsAborted(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 10
	}
	return false
}

// OutOfRange The operation was attempted past the valid range.  E.g., seeking or
// reading past end-of-file.
// HTTP Mapping: 400 Bad Request
func OutOfRange(id, format string, a ...interface{}) error {
	return &Error{
		Code:    11,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsOutOfRange determines if err is an error which indicates a out of range error.
// It supports wrapped errors.
func IsOutOfRange(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 11
	}
	return false
}

// Unimplemented The operation is not implemented or is not supported/enabled in this service.
// HTTP Mapping: 501 Not Implemented
func Unimplemented(id, format string, a ...interface{}) error {
	return &Error{
		Code:    12,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsUnimplemented determines if err is an error which indicates a unimplemented error.
// It supports wrapped errors.
func IsUnimplemented(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 12
	}
	return false
}

// Internal This means that some invariants expected by the
// underlying system have been broken.  This error code is reserved
// for serious errors.
//
// HTTP Mapping: 500 Internal Server Error
func Internal(id, format string, a ...interface{}) error {
	return &Error{
		Code:    13,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsInternal determines if err is an error which indicates an internal server error.
// It supports wrapped errors.
func IsInternal(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 13
	}
	return false
}

// Unavailable The service is currently unavailable.
// HTTP Mapping: 503 Service Unavailable
func Unavailable(id, format string, a ...interface{}) error {
	return &Error{
		Code:    14,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsUnavailable determines if err is an error which indicates a unavailable error.
// It supports wrapped errors.
func IsUnavailable(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 14
	}
	return false
}

// DataLoss Unrecoverable data loss or corruption.
// HTTP Mapping: 500 Internal Server Error
func DataLoss(id, format string, a ...interface{}) error {
	return &Error{
		Code:    15,
		Id:      id,
		Message: fmt.Sprintf(format, a...),
	}
}

// IsDataLoss determines if err is an error which indicates a data loss error.
// It supports wrapped errors.
func IsDataLoss(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 15
	}
	return false
}

// IsUnauthorized determines if err is an error which indicates a unauthorized error.
// It supports wrapped errors.
func IsUnauthorized(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 16
	}
	return false
}
