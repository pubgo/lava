package errors

import (
	"fmt"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/proto/errorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New generates a custom error.
func New(reason string) *Error {
	return &Error{
		reason:  reason,
		version: version.Version(),
		service: version.Project(),
	}
}

func NewWithBizCode(code string, reason string) *Error {
	return &Error{
		bizCode: code,
		reason:  reason,
		version: version.Version(),
		service: version.Project(),
	}
}

type Error struct {
	version   string
	service   string
	bizCode   string
	operation string
	code      uint32
	reason    string
	err       error
	tags      map[string]string
}

// GRPCStatus 实现 grpc status 的GRPCStatus接口
func (e Error) GRPCStatus() *status.Status {
	if e.err == nil {
		return nil
	}

	return assert.Must1(status.New(codes.Code(e.code), e.err.Error()).WithDetails(e.Proto()))
}

func (e Error) Proto() *errorpb.Error {
	if e.err == nil {
		return nil
	}

	return &errorpb.Error{
		Code:      e.code,
		BizCode:   e.bizCode,
		Service:   e.service,
		Version:   e.version,
		Operation: e.operation,
		Reason:    e.reason,
		Tags:      e.tags,
		ErrMsg:    e.err.Error(),
		ErrDetail: fmt.Sprintf("%+v", e.err),
	}
}

// HTTPStatus returns the Status represented by se.
func (e Error) HTTPStatus() int {
	return GrpcCodeToHTTP(codes.Code(e.code))
}

func (e Error) Ok() bool {
	return e.err == nil || e.code == 0
}

// Is matches each error in the chain with the target value.
func (e Error) Is(err error) bool {
	if err == nil {
		return false
	}

	if err1, ok := err.(*Error); ok {
		return e.code == err1.code && e.reason == err1.reason
	}

	return false
}

func (e Error) As(target any) bool {
	switch x := target.(type) {
	case nil:
		return false
	case *Error:
		*x = e
		return true
	case **Error:
		*x = &e
		return true
	}

	return false
}

func (e Error) Tags(tags map[string]string) Error {
	if tags == nil || len(tags) == 0 {
		return e
	}

	for k, v := range e.tags {
		tags[k] = v
	}
	e.tags = tags
	return e
}

func (e Error) BizCode(code string) Error {
	e.bizCode = code
	return e
}

func (e Error) Operation(operation string) Error {
	e.operation = operation
	return e
}

func (e Error) Reason(reason string) Error {
	e.reason = reason
	return e
}

func (e Error) Error() string {
	if e.err == nil {
		return ""
	}

	return fmt.Sprintf("version=%q service=%q operation=%q code=%d biz_code=%q status=%s err_msg=%q err_detail=%+v",
		e.version, e.service, e.operation, e.code, e.bizCode, e.reason, e.err.Error(), e.err)
}

func (e Error) Err(err error) Error {
	if err != nil {
		e.err = err
	}
	return e
}

func (e Error) Status(code codes.Code) Error {
	e.code = uint32(code)
	return e
}
