package version

import (
	"runtime"

	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/version"
)

func GetVer() map[string]interface{} {
	return map[string]interface{}{
		"device_id":     runenv.DeviceId,
		"project":       runenv.Project,
		"data":          version.Data,
		"build_time":    version.BuildTime,
		"version":       version.Version,
		"tag":           version.Tag,
		"commit_id":     version.CommitID,
		"domain":        version.Domain,
		"go_root":       runtime.GOROOT(),
		"go_arch":       runtime.GOARCH,
		"go_os":         runtime.GOOS,
		"go_version":    runtime.Version(),
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
	}
}
