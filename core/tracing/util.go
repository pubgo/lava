package tracing

import (
	"context"
	"runtime"

	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	KeyErrMsg = "err_msg"
)

// SetIfErr add error info
func SetIfErr(span oteltrace.Span, err error) {
	if span == nil || err == nil {
		return
	}

	span.RecordError(err)
}

// SetIfCtxErr record context error
func SetIfCtxErr(span oteltrace.Span, ctx context.Context) {
	if span == nil || ctx == nil {
		return
	}

	err := ctx.Err()
	if err == nil {
		return
	}

	SetIfErr(span, err)
	span.SpanContext()
}

func queueSize() int {
	const min = 1000
	const max = 16000

	n := (runtime.GOMAXPROCS(0) / 2) * 1000
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
