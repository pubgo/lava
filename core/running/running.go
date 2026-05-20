package running

import (
	"fmt"
	"os"
	rt "runtime"
	"strings"

	semver "github.com/hashicorp/go-version"
	"github.com/projectdiscovery/machineid"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/debugs"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/netutil"
	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/pubgo/redant"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

var (
	Env      = redant.StringOf(lo.ToPtr("dev"))
	Debug    = redant.BoolOf(lo.ToPtr(false))
	HttpPort = redant.Int64Of(lo.ToPtr(int64(8080)))
	GrpcPort = redant.Int64Of(lo.ToPtr(int64(50051)))
	Project  = version.Project

	InstanceID = xid.New().String()
	DeviceID   = InstanceID

	Version  = version.Version
	CommitID = version.CommitID

	Pwd = assert.Exit1(os.Getwd())

	LocalIP = netutil.GetLocalIP()

	Hostname = strutil.FirstFnNotEmpty(
		func() string { return env.Get("HOSTNAME") },
		func() string { return assert.Exit1(os.Hostname()) },
	)

	Namespace = strutil.FirstFnNotEmpty(
		func() string { return env.Get("NAMESPACE") },
		func() string { return env.Get("POD_NAMESPACE") },
		func() string {
			file := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
			if pathutil.IsNotExist(file) {
				return ""
			}

			return strings.TrimSpace(string(assert.Exit1(os.ReadFile(file))))
		},
	)

	Domain = version.Domain

	DebugFlag = redant.Option{
		Flag:        "debug",
		Description: "enable debug mode",
		Value:       Debug,
		Default:     Debug.String(),
		Envs:        []string{env.Key("enable_debug"), env.Key("debug")},
		Action: func(val pflag.Value) error {
			env.Set("enable_debug", val.String())
			env.Set("debug", val.String())
			return debugs.Enabled.Set(val.String())
		},
	}

	EnvFlag = redant.Option{
		Flag:        "runenv",
		Description: "running env, dev,test,stage,prod",
		Value:       Env,
		Default:     Env.String(),
		Envs:        []string{env.Key("env"), env.Key("runenv")},
		Action: func(val pflag.Value) error {
			env.Set("env", val.String())
			env.Set("runenv", val.String())
			return nil
		},
	}

	GrpcPortFlag = redant.Option{
		Flag:        "grpc-port",
		Description: "service grpc port",
		Value:       GrpcPort,
		Default:     GrpcPort.String(),
		Envs:        []string{env.Key("server_grpc_port")},
		Action: func(val pflag.Value) error {
			env.Set("server_grpc_port", val.String())
			return nil
		},
	}

	HttpPortFlag = redant.Option{
		Flag:        "http-port",
		Description: "service http port",
		Value:       HttpPort,
		Default:     HttpPort.String(),
		Envs:        []string{env.Key("server_http_port")},
		Action: func(val pflag.Value) error {
			env.Set("server_http_port", val.String())
			return nil
		},
	}
)

func init() {
	id, err := machineid.ID()
	if err == nil {
		DeviceID = id
	}
}

func GetSysInfo() map[string]string {
	return map[string]string{
		"main_path":     version.MainPath(),
		"grpc_port":     GrpcPort.String(),
		"http_port":     HttpPort.String(),
		"debug":         Debug.String(),
		"cur_dir":       Pwd,
		"local_ip":      LocalIP,
		"namespace":     Namespace,
		"instance_id":   InstanceID,
		"device_id":     DeviceID,
		"project":       Project(),
		"hostname":      Hostname,
		"build_time":    version.BuildTime(),
		"version":       Version(),
		"domain":        Domain(),
		"commit_id":     CommitID(),
		"go_root":       env.Get("GOROOT"),
		"go_arch":       rt.GOARCH,
		"go_os":         rt.GOOS,
		"go_version":    rt.Version(),
		"num_cpu":       fmt.Sprintf("%v", rt.NumCPU()),
		"num_goroutine": fmt.Sprintf("%v", rt.NumGoroutine()),
	}
}

func CheckVersion() {
	defer recovery.Exit()
	assert.If(version.Project() == "", "project is null")
	assert.If(version.Version() == "", "version is null")
	assert.If(version.CommitID() == "", "commitID is null")
	assert.If(version.BuildTime() == "", "buildTime is null")
	assert.MustFn(func() error {
		_, err := semver.NewVersion(version.Version())
		if err != nil {
			return fmt.Errorf("version(%s) error: %w", version.Version(), err)
		}
		return nil
	})
}
