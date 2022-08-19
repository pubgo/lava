package logging

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/config"
)

func init() {
	dix.Provider(func() ExtLog { return func(log *Logger) {} })
	dix.Provider(func(c config.Config, logs []ExtLog) *Logger {
		var log = New(c)
		for i := range logs {
			logs[i](log)
		}
		return log
	})
}
