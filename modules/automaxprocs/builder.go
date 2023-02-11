package automaxprocs

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"go.uber.org/automaxprocs/maxprocs"
)

// doc: https://github.com/KimMachineGun/automemlimit

const name = "automaxprocs"

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		ff := log.GetLogger(name).WithCallerSkip(2).Info().Msgf
		assert.Must1(maxprocs.Set(maxprocs.Logger(ff)))
	}
}
