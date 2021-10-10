package logger

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lug/pkg/fastrand"
)

var Discard = zap.NewNop()

func On(fn func(log *zap.Logger)) *zap.Logger {
	xerror.Exit(dix.Provider(fn))
	return zap.L()
}

// Probability 根据概率获取真实的logger
func Probability(prob float64) *zap.Logger {
	if fastrand.Probability(prob) {
		return Discard
	}
	return zap.L()
}
