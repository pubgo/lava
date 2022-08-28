package logging

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/config"
)

func init() {
	di.Provide(func() ExtLog { return func(log *Logger) {} })
	di.Provide(func(c config.Config, logs []ExtLog) *Logger {
		var log = New(c)
		for i := range logs {
			logs[i](log)
		}
		return log
	})
}
