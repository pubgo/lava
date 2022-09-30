package tracing

import (
	"encoding/binary"
	"time"

	"github.com/pubgo/funk/fastrand"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
)

var (
	randomInitSequence = int32(fastrand.Uint32())
	sequence           = atomic.NewInt32(randomInitSequence)
)

// NewIDs creates and returns a new trace and span ID.
func NewIDs() (traceID trace.TraceID, spanID trace.SpanID) {
	return NewTraceID(), NewSpanID()
}

// NewTraceID creates and returns a trace ID.
func NewTraceID() (traceID trace.TraceID) {
	binary.LittleEndian.PutUint64(traceID[:], uint64(time.Now().UnixNano()))
	binary.LittleEndian.PutUint32(traceID[8:], uint32(sequence.Add(1)))
	copy(traceID[12:], fastrand.Bytes(4))
	return
}

// NewSpanID creates and returns a span ID.
func NewSpanID() (spanID trace.SpanID) {
	binary.LittleEndian.PutUint64(spanID[:], uint64(time.Now().UnixNano()/1e3))
	copy(spanID[4:], fastrand.Bytes(4))
	return
}
