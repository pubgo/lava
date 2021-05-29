package metric

import (
	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"

	"context"
	"fmt"
	"time"
)

func init() {
	var name = "health"
	xerror.Panic(CreateCounter(name, typex.StrOf("version", "build_time", "commit_id"), CounterOpts{
		Help: fmt.Sprintf("%s health check", app.Project),
	}))

	var tags = Tags{"version": version.Version, "build_time": version.BuildTime, "commit_id": version.CommitID}

	_ = fx.Go(func(_ context.Context) {
		for range time.Tick(time.Second) {
			var err = Count(name, 1.0, tags)
			if err != nil {
				xlog.ErrorM("health check error", xlog.M{"err": err, "version": tags})
			}
		}
	})
}