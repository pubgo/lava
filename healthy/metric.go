package healthy

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/metric"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/version"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	entry.AfterStart(func() {
		var tags = metric.Tags{"version": version.Version, "build_time": version.BuildTime, "commit_id": version.CommitID}

		xerror.Panic(metric.CreateCounter(Name, typex.StrOf("version", "build_time", "commit_id"), metric.CounterOpts{
			Help: fmt.Sprintf("%s health check", runenv.Project),
		}))

		_ = fx.Go(func(_ context.Context) {
			for range time.Tick(time.Second) {
				if err := metric.Count(Name, 1.0, tags); err != nil {
					xlog.Error("health check", xlog.M{"err": err, "tags": tags})
				}
			}
		})
	})
}
