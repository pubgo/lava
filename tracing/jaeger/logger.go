package jaeger

import (
	"fmt"

	"github.com/pubgo/xlog"
	"github.com/uber/jaeger-client-go"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
)

var _ jaegerLog.Logger = (*logger)(nil)

type logger struct {
	logs xlog.Xlog
}

func (l logger) Debugf(msg string, args ...interface{}) {
	l.logs.Debugf(msg, args...)
}

func (l logger) Error(msg string) {
	l.logs.Error(msg)
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logs.Infof(msg, args...)
}

func (l logger) Write(b []byte) (n int, err error) {
	l.logs.Info(string(b))
	return len(b), nil
}

// Report implements Report() method of Reporter by logging the span to the logger.
func (l *logger) Report(span *jaeger.Span) {
	for _, logs := range span.Logs() {
		var fields []interface{}
		for i := range logs.Fields {
			fields = append(fields, logs.Fields[i].String())
		}

		l.logs.Info(
			fmt.Sprintf("Reporting span %s %+v", span.OperationName(), span),
			zap.Time("StartTime", span.StartTime()),
			zap.Any("tags", span.Tags()),
			zap.Duration("Duration", span.Duration()),
			zap.Any("fields", fields),
		)
	}
}

// Close implements Close() method of Reporter by doing nothing.
func (l *logger) Close() {}
