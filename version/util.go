package version

import (
	"github.com/pubgo/lava/core/runmode"
	rt "runtime"
)

func GetVersion() map[string]interface{} {
	return map[string]interface{}{
		"namespace":     runmode.Namespace,
		"app_id":        runmode.InstanceID,
		"device_id":     runmode.DeviceID,
		"project":       runmode.Project,
		"data":          Data,
		"build_time":    BuildTime,
		"version":       Version,
		"tag":           Tag,
		"commit_id":     CommitID,
		"domain":        Domain,
		"go_root":       rt.GOROOT(),
		"go_arch":       rt.GOARCH,
		"go_os":         rt.GOOS,
		"go_version":    rt.Version(),
		"num_cpu":       rt.NumCPU(),
		"num_goroutine": rt.NumGoroutine(),
	}
}
