package logging

import (
	"go.uber.org/zap"
)

const Name = "logger"

type Logger = zap.Logger

type Fields = []zap.Field

// L global zap log
func L() *zap.Logger {
	return zap.L()
}

// S global zap sugared log
func S() *zap.SugaredLogger {
	return zap.S()
}
