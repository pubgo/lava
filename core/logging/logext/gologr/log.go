package gologr

import (
	"fmt"
	"slices"

	"github.com/go-logr/logr"
	"github.com/pubgo/funk/log"
	"github.com/rs/zerolog"
)

var (
	RenderArgsHook   = defaultRender
	RenderValuesHook = defaultRender
)

// LogSink implements logr.LogSink and logr.CallDepthLogSink.
type LogSink struct {
	l             log.Logger
	keysAndValues []any
}

var (
	_ logr.LogSink          = &LogSink{}
	_ logr.CallDepthLogSink = &LogSink{}
)

func NewSink(l log.Logger) *LogSink {
	return &LogSink{l: l}
}

func (ls *LogSink) Init(ri logr.RuntimeInfo) {
	ls.l = ls.l.WithCallerSkip(ri.CallDepth + 2)
}

func (ls *LogSink) Enabled(level int) bool { return true }

func (ls *LogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	ls.msg(ls.l.Info(), msg, keysAndValues)
}

func (ls *LogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	ls.msg(ls.l.Err(err), msg, keysAndValues)
}

func (ls *LogSink) msg(e *zerolog.Event, msg string, keysAndValues []interface{}) {
	if e == nil {
		return
	}

	if RenderArgsHook != nil {
		keysAndValues = RenderArgsHook(keysAndValues)
	}

	e = e.Fields(keysAndValues)
	e.Msg(msg)
}

func (ls *LogSink) copy() *LogSink {
	return &LogSink{
		l:             ls.l,
		keysAndValues: slices.Clone(ls.keysAndValues),
	}
}

func (ls *LogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	if len(keysAndValues) == 0 {
		return ls
	}

	if RenderValuesHook != nil {
		keysAndValues = RenderValuesHook(keysAndValues)
	}

	ll := ls.copy()
	ll.keysAndValues = append(ll.keysAndValues, keysAndValues...)

	return ll
}

func (ls *LogSink) WithName(name string) logr.LogSink {
	ll := ls.copy()
	ll.l = ll.l.WithName(name)
	return ll
}

func (ls *LogSink) WithCallDepth(depth int) logr.LogSink {
	ll := ls.copy()
	ll.l = ll.l.WithCallerSkip(depth)
	return ll
}

func defaultRender(keysAndValues []interface{}) []interface{} {
	for i, n := 1, len(keysAndValues); i < n; i += 2 {
		value := keysAndValues[i]
		switch v := value.(type) {
		case logr.Marshaler:
			keysAndValues[i] = v.MarshalLog()
		case fmt.Stringer:
			keysAndValues[i] = v.String()
		}
	}
	return keysAndValues
}
