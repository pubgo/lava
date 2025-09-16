package slog

import (
	"log/slog"
	"testing"

	"github.com/pubgo/funk/v2/log"
)

func TestName(t *testing.T) {
	SetLogger(log.GetLogger("testing"))

	slog.Info("hello")
	slog.With(slog.String("name", "abc")).Info("hello")
	slog.With(slog.String("name", "abc")).Info("hello", "data-map", map[string]any{
		"abc": 1,
		"hello": map[string]any{
			"world": 2,
		},
	})
}
