package runmode

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/utils"
)

// 默认的全局配置
var (
	Domain  = "lava"
	Block   = true
	Project = env.Get("app_name", "service_name", "project_name")
	Level   = "debug"

	// DeviceID 主机设备ID
	DeviceID = xerror.ExitErr(machineid.ID())

	// InstanceID service id
	InstanceID = uuid.New().String()

	Signal os.Signal = syscall.Signal(0)

	// Pwd 当前目录
	Pwd = xerror.ExitErr(os.Getwd()).(string)

	// Hostname 主机名
	Hostname = utils.FirstFnNotEmpty(
		func() string { return os.Getenv("HOSTNAME") },
		func() string {
			flagx.ExampleFmt()
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
