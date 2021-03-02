package version

import (
	"runtime"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xerror"
)

var GoVersion = runtime.Version()
var GoPath = ""
var GoROOT = ""
var CommitID = ""
var Project = ""

func init() {
	xerror.ExitErr(ver.NewVersion(Version))
	tracelog.Watch("golug_version", func() interface{} {
		return map[string]interface{}{
			"build_time": BuildTime,
			"version":    Version,
			"go_version": GoVersion,
			"go_path":    GoPath,
			"go_root":    GoROOT,
			"commit_id":  CommitID,
			"project":    Project,
		}
	})
}
