package errors

import (
	"context"
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	errorV1 "github.com/pubgo/lava/pkg/proto/errors/v1"
)

// New generates a custom error.
func New(code string) *Error {
	return &Error{err: &errorV1.Error{Code: code, Tags: make(map[string]string)}}
}

type Error struct {
	err *errorV1.Error
}

// GRPCStatus 实现grpc status的GRPCStatus接口
func (e *Error) GRPCStatus() *status.Status {
	var dt = assert.Must1(json.Marshal(e.err))
	return status.New(codes.Code(e.err.Status), string(dt))
}

// HTTPStatus returns the Status represented by se.
func (e *Error) HTTPStatus() int {
	return GrpcCodeToHTTP(codes.Code(e.err.Status))
}

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if err == nil {
		return false
	}

	if err1, ok := err.(*Error); ok {
		return e.err.Code == err1.err.Code && e.err.Status == err1.err.Status
	}

	return false
}

func (e *Error) As(target any) bool {
	switch x := target.(type) {
	case nil:
		return false
	case *Error:
		*x = *e
		return true
	case **Error:
		*x = e
		return true
	}

	t1 := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if err, ok := t1.(*Error); ok {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
		return true
	}

	return false
}

// Ctx get some metadata from ctx
func (e *Error) Ctx(ctx context.Context) *Error {
	return e
}

func (e *Error) Msg(msg string, args ...interface{}) *Error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Msg = fmt.Sprintf(msg, args...)
	return &Error{err: ee}
}

func (e *Error) Tag(k, v string) *Error {
	if k == "" {
		return e
	}

	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Tags[k] = v
	return &Error{err: ee}
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code=%s status=%d msg=%s tags=%v", e.err.Code, e.err.Status, e.err.ErrMsg, e.err.Tags)
}

func (e *Error) Err(err error) *Error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	if err != nil {
		ee.ErrMsg = err.Error()
		ee.ErrDetail = fmt.Sprintf("%#v", err)
	}
	return &Error{err: ee}
}

func (e *Error) Status(code codes.Code) *Error {
	var ee = proto.Clone(e.err).(*errorV1.Error)
	ee.Status = uint32(code)
	return &Error{err: ee}
}
