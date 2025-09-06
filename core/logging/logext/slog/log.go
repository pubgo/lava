package slog

import (
	"context"
	"log/slog"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/v2/core/logging"
	"github.com/rs/zerolog"
	slogcommon "github.com/samber/slog-common"
)

var evt = log.NewEvent().Str("ext", "slog")
var logLevels = map[slog.Level]zerolog.Level{
	slog.LevelDebug: zerolog.DebugLevel,
	slog.LevelInfo:  zerolog.InfoLevel,
	slog.LevelWarn:  zerolog.WarnLevel,
	slog.LevelError: zerolog.ErrorLevel,
}

func init() {
	logging.Register("slog", SetLogger)
}

func SetLogger(logger log.Logger) {
	slog.SetDefault(slog.New(&std{l: logger.WithEvent(evt)}))
}

var _ slog.Handler = (*std)(nil)

type std struct {
	l log.Logger
}

func (s *std) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (s *std) Handle(ctx context.Context, r slog.Record) error {
	logger := s.l.WithCallerSkip(3).WithLevel(logLevels[r.Level])
	event := logger.Info(ctx)
	switch r.Level {
	case slog.LevelDebug:
		event = logger.Debug(ctx)
	case slog.LevelInfo:
	case slog.LevelWarn:
		event = logger.Warn(ctx)
	case slog.LevelError:
		event = logger.Error(ctx)
	}

	if !r.Time.IsZero() {
		event.Time(zerolog.TimestampFieldName, r.Time)
	}

	r.Attrs(func(attr slog.Attr) bool {
		event.Any(attr.Key, attr.Value.Any())
		return true
	})

	event.Msg(r.Message)
	return nil
}

func (s *std) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &std{l: s.l.WithFields(slogcommon.AttrsToMap(attrs...))}
}

func (s *std) WithGroup(name string) slog.Handler {
	return &std{l: s.l.WithName(name)}
}
