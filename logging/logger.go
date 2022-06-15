package logging

import (
	"go.uber.org/zap"
)

const Name = "logging"

var global *Logger

type ExtLog func(logger *Logger)
type Logger = zap.Logger

// L global zap log
func L() *zap.Logger {
	return zap.L()
}

// S global zap sugared log
func S() *zap.SugaredLogger {
	return zap.S()
}
