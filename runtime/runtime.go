package runtime

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/version"
)

// 默认的全局配置
var (
	Domain  = version.Domain
	Block   = true
	Trace   = false
	Addr    = ":8080"
	Project = env.Get("app_name", "service_name", "project_name")
	Level   = "debug"
	Mode    = RunModeLocal

	// DeviceID 主机设备ID
	DeviceID = xerror.ExitErr(machineid.ID())

	// AppID service id
	AppID = uuid.New().String()

	Signal os.Signal = syscall.Signal(0)

	// Pwd 当前目录
	Pwd = xerror.ExitErr(os.Getwd()).(string)

	// Hostname 主机名
	Hostname = utils.FirstFnNotEmpty(
		func() string { return os.Getenv("HOSTNAME") },
		func() string {
			var h, err = os.Hostname()
			xerror.Exit(err)
			return h
		},
	)

	// Namespace 命名空间
	Namespace = utils.FirstFnNotEmpty(
		func() string { return os.Getenv("NAMESPACE") },
		func() string { return os.Getenv("POD_NAMESPACE") },
		func() string {
			if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
				if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
					return ns
				}
			}
			return ""
		},
	)
)
