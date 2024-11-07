package grpcs

import (
	"net/url"
	"strings"

	"github.com/ettle/strcase"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/proto/errorpb"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// serviceFromMethod returns the service
// /service.Foo/Bar => service.Foo
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	return strings.Split(strings.Trim(m, "/"), "/")[0]
}

type DefaultQueryParser struct{}

// Parse populates "values" into "msg".
// A value is ignored if its key starts with one of the elements in "filter".
func (*DefaultQueryParser) Parse(msg proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	for key, v := range values {
		if len(v) == 0 {
			delete(values, key)
			continue
		}

		if len(v) == 1 && v[0] == "" {
			delete(values, key)
			continue
		}
	}

	return new(runtime.DefaultQueryParser).Parse(msg, values, filter)
}

func handlerHttpErr(err error) error {
	if err == nil {
		return nil
	}

	var errPb *errorpb.ErrCode
	if errors.As(err, &errPb) {
		if errPb.Message == "" {
			errPb.Message = err.Error()
		}
	}

	if errPb == nil {
		sts, ok := status.FromError(err)
		if ok && sts != nil {
			if len(sts.Details()) > 0 {
				errDetail := sts.Details()[0]
				if code, ok := errDetail.(*errorpb.Error); ok {
					errPb = code.Code
				}

				if code, ok := errDetail.(*errorpb.ErrCode); ok {
					errPb = code
				}
			} else {
				errPb = &errorpb.ErrCode{
					Message:    sts.Message(),
					Code:       int32(errorpb.Code(sts.Code())),
					StatusCode: errorpb.Code(sts.Code()),
					Name:       "code." + strcase.ToSnake(errorpb.Code(sts.Code()).String()),
					Details:    sts.Proto().Details,
				}
			}
		}
	}

	if errPb == nil {
		errPb = &errorpb.ErrCode{
			Message:    err.Error(),
			StatusCode: errorpb.Code_Internal,
			Code:       int32(errorpb.Code_Internal),
			Name:       "code.internal",
		}
	}

	// skip error
	if errPb.StatusCode == errorpb.Code_OK {
		return nil
	}

	return errors.NewCodeErr(errPb, errors.ParseErrToPb(err))
}
