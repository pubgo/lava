package version

import (
	"github.com/pubgo/lava/version"
	"runtime"

	"github.com/pubgo/lava/runenv"
)

func GetVer() map[string]interface{} {
	return map[string]interface{}{
		"data":       version.Data,
		"build_time": version.BuildTime,
		"version":    version.Version,
		"go_root":    runtime.GOROOT(),
		"go_arch":    runtime.GOARCH,
		"go_os":      runtime.GOOS,
		"go_version": runtime.Version(),
		"commit_id":  version.CommitID,
		"project":    runenv.Project,
		"domain":     version.Domain,
	}
}
