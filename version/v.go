package version

import (
	"runtime"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/golug/golug_version"
	"github.com/pubgo/xerror"
)

var GoVersion = runtime.Version()
var GoPath = ""
var GoROOT = ""
var CommitID = ""
var Project = ""

func init() {
	xerror.ExitErr(ver.NewVersion(Version))
	golug_version.Register("golug_version", golug_version.M{
		"build_time": BuildTime,
		"version":    Version,
		"go_version": GoVersion,
		"go_path":    GoPath,
		"go_root":    GoROOT,
		"commit_id":  CommitID,
		"project":    Project,
	})
}
