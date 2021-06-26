package metric

import (
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/runenv"
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
		Help: fmt.Sprintf("%s health check", runenv.Project),
	}))

	var tags = Tags{"version": version.Version, "build_time": version.BuildTime, "commit_id": version.CommitID}

	_ = fx.Go(func(_ context.Context) {
		for range time.Tick(time.Second) {
			if err := Count(name, 1.0, tags); err != nil {
				logs.Error("health check", xlog.M{"err": err, "tags": tags})
			}
		}
	})
}
