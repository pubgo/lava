package runmode

import (
	rt "runtime"

	"github.com/pubgo/lava/version"
)

func GetVersion() map[string]interface{} {
	return map[string]interface{}{
		"grpc_port":     GrpcPort,
		"http_post":     HttpPort,
		"debug":         IsDebug,
		"pwd":           Pwd,
		"namespace":     Namespace,
		"instance_id":   InstanceID,
		"device_id":     DeviceID,
		"project":       Project,
		"hostname":      Hostname,
		"build_time":    version.BuildTime(),
		"version":       Version,
		"commit_id":     CommitID,
		"domain":        version.Domain(),
		"go_root":       rt.GOROOT(),
		"go_arch":       rt.GOARCH,
		"go_os":         rt.GOOS,
		"go_version":    rt.Version(),
		"num_cpu":       rt.NumCPU(),
		"num_goroutine": rt.NumGoroutine(),
	}
}
