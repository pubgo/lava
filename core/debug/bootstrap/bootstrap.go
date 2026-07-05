// Package bootstrap registers built-in debug routes on the global debug App.
package bootstrap

import (
	"sync"

	"github.com/pubgo/lava/v2/core/debug/configview"
	debugconsole "github.com/pubgo/lava/v2/core/debug/debug"
	"github.com/pubgo/lava/v2/core/debug/featurehttp"
	"github.com/pubgo/lava/v2/core/debug/goroutine"
	"github.com/pubgo/lava/v2/core/debug/healthy"
	"github.com/pubgo/lava/v2/core/debug/loglevel"
	"github.com/pubgo/lava/v2/core/debug/pprof"
	"github.com/pubgo/lava/v2/core/debug/process"
	"github.com/pubgo/lava/v2/core/debug/ratelimit"
	"github.com/pubgo/lava/v2/core/debug/runtime"
	"github.com/pubgo/lava/v2/core/debug/statsviz"
	"github.com/pubgo/lava/v2/core/debug/sysinfo"
	"github.com/pubgo/lava/v2/core/debug/trace"
	"github.com/pubgo/lava/v2/core/debug/vars"
	"github.com/pubgo/lava/v2/core/debug/version"
)

var registerOnce sync.Once

// RegisterAll registers built-in debug routes and middleware on debug.App().
func RegisterAll() {
	registerOnce.Do(func() {
		debugconsole.Register()
		ratelimit.Register()
		configview.Register()
		featurehttp.Register()
		goroutine.Register()
		healthy.Register()
		loglevel.Register()
		pprof.Register()
		process.Register()
		runtime.Register()
		statsviz.Register()
		sysinfo.Register()
		trace.Register()
		vars.Register()
		version.Register()
	})
}
