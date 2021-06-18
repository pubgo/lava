package version

import (
	"runtime"

	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/xerror"
)

var GoVersion = runtime.Version()
var GoPath = ""
var GoROOT = ""
var CommitID = ""
var Project = ""
var BuildTime = "2021-03-20 16:52:09"
var Version = "v0.0.19"
var Data = ""

func GetVer() map[string]interface{} {
	return map[string]interface{}{
		"data":       Data,
		"build_time": BuildTime,
		"version":    Version,
		"go_version": GoVersion,
		"go_path":    GoPath,
		"go_root":    GoROOT,
		"commit_id":  CommitID,
		"project":    Project,
	}
}

func init() {
	xerror.ExitErr(ver.NewVersion(Version))
	vars.Watch("version", func() interface{} { return GetVer() })
}
