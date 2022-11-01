package errors

import (
	"fmt"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/pkg/proto/errorpb"
	"github.com/pubgo/lava/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New generates a custom error.
func New(reason string) *Error {
	return &Error{
		reason: reason,
		tags: map[string]string{
			"project": version.Project(),
			"version": version.Version(),
		},
	}
}

type Error struct {
	err    error
	code   uint32
	reason string
	tags   map[string]string
}

// GRPCStatus 实现 grpc status 的GRPCStatus接口
func (e Error) GRPCStatus() *status.Status {
	if e.err == nil || e.code == 0 {
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
		Reason:    e.reason,
		Tags:      e.tags,
		ErrMsg:    e.err.Error(),
		ErrDetail: fmt.Sprintf("%#v", e.err),
	}
}

// HTTPStatus returns the Status represented by se.
func (e Error) HTTPStatus() int {
	return GrpcCodeToHTTP(codes.Code(e.code))
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

func (e Error) Error() string {
	if e.err == nil {
		return ""
	}

	return fmt.Sprintf("version=%q project=%s code=%d status=%s err_msg=%q tags=%v",
		version.Version(), version.Project(), e.code, e.reason, e.err.Error(), e.tags)
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
