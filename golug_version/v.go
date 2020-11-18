package golug_version

import (
	"runtime"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/golug/version"
	"github.com/pubgo/xerror"
)

var BuildTime = version.BuildTime
var Version = version.Version
var GoVersion = runtime.Version()
var GoPath = ""
var GoROOT = ""
var CommitID = ""
var Project = ""

func init() {
	if Version == "" {
		Version = "v0.0.1"
	}

	xerror.ExitErr(ver.NewVersion(Version))
	xerror.Exit(Register("golug_version", M{
		"build_time": BuildTime,
		"version":    Version,
		"go_version": GoVersion,
		"go_path":    GoPath,
		"go_root":    GoROOT,
		"commit_id":  CommitID,
		"project":    Project,
	}))
}
