package runmode

import (
	rt "runtime"

	"github.com/pubgo/lava/version"
)

func GetVersion() map[string]interface{} {
	return map[string]interface{}{
		"namespace":     Namespace,
		"app_id":        InstanceID,
		"device_id":     DeviceID,
		"project":       Project,
		"data":          version.Data,
		"build_time":    version.BuildTime,
		"version":       version.Version(),
		"tag":           version.Tag,
		"commit_id":     version.CommitID,
		"domain":        version.Domain,
		"go_root":       rt.GOROOT(),
		"go_arch":       rt.GOARCH,
		"go_os":         rt.GOOS,
		"go_version":    rt.Version(),
		"num_cpu":       rt.NumCPU(),
		"num_goroutine": rt.NumGoroutine(),
	}
}
