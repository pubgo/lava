// Package running 提供进程级运行态信息与全局 CLI flag 定义。
//
// 包含环境名（dev/test/stage/prod）、debug 开关、HTTP/gRPC 端口、
// 实例 ID、主机名、K8s namespace 等，供日志、debug 端点和各命令共享。
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
	// Env 是当前运行环境，默认 "dev"。
	Env = redant.StringOf(lo.ToPtr("dev"))
	// Debug 是否启用 debug 模式。
	Debug = redant.BoolOf(lo.ToPtr(false))
	// HttpPort 是默认 HTTP 监听端口。
	HttpPort = redant.Int64Of(lo.ToPtr(int64(8080)))
	// GrpcPort 是默认 gRPC 监听端口。
	GrpcPort = redant.Int64Of(lo.ToPtr(int64(50051)))
	// Project 返回项目名称（来自 buildinfo）。
	Project = version.Project

	// InstanceID 是当前进程实例的唯一 ID，进程启动时生成。
	InstanceID = xid.New().String()
	// DeviceID 是设备级唯一标识，优先使用 machineid，失败时回退为 InstanceID。
	DeviceID = InstanceID

	Version  = version.Version
	CommitID = version.CommitID

	// Pwd 是进程启动时的工作目录。
	Pwd = assert.Exit1(os.Getwd())

	LocalIP = netutil.GetLocalIP()

	Hostname = strutil.FirstFnNotEmpty(
		func() string { return env.Get("HOSTNAME") },
		func() string { return assert.Exit1(os.Hostname()) },
	)

	// Namespace 是 K8s 命名空间，依次从环境变量或 serviceaccount 文件读取。
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

// GetSysInfo 返回当前进程的系统与构建信息快照，供 debug/version 命令使用。
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

// CheckVersion 校验 buildinfo 中的项目名、版本号等必填字段，不合法时 panic。
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
