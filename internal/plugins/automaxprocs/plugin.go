package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/internal/logz"
)

func init() {
	const name = "automaxprocs"

	// https://pkg.go.dev/go.uber.org/automaxprocs
	// Automatically set GOMAXPROCS to match Linux container CPU quota.
	logz.On(func(_ *logz.Log) {
		var logs = logz.New(name).DepthS(1)
		xerror.ExitErr(maxprocs.Set(maxprocs.Logger(logs.Infof))).(func())()
	})
}
