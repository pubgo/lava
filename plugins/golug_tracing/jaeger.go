package golug_tracing

import (
	"context"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"
)

const (
	// error keys
	KeyErrorMessage        = "err_msg"
	KeyContextErrorMessage = "ctx_err_msg"
)

func SetIfErr(span opentracing.Span, err error) {
	if span == nil || err == nil {
		return
	}

	ext.Error.Set(span, true)
	span.SetTag(KeyErrorMessage, err.Error())
}

func SetIfCtxErr(span opentracing.Span, ctx context.Context) {
	if span == nil || ctx == nil {
		return
	}

	if err := ctx.Err(); err != nil {
		ext.Error.Set(span, true)
		span.SetTag(KeyContextErrorMessage, err.Error())
	}
}
