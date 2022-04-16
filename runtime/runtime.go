package runtime

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	dir "github.com/mitchellh/go-homedir"
	"github.com/pubgo/xerror"
	"k8s.io/client-go/util/homedir"

	"github.com/pubgo/lava/internal/envs"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/version"
)

// 默认的全局配置
var (
	Domain  = version.Domain
	Block   = true
	Trace   = false
	Addr    = ":8080"
	Project = "lava"
	Level   = "debug"
	Mode    = RunModeDev

	// DeviceID 主机设备ID
	DeviceID = xerror.ExitErr(machineid.ID())

	// AppID service id
	AppID = uuid.New().String()

	Signal os.Signal = syscall.Signal(0)

	// Pwd 当前目录
	Pwd = xerror.ExitErr(os.Getwd()).(string)

	// Hostname 主机名
	Hostname = utils.FirstNotEmpty(
		func() string { return os.Getenv("HOSTNAME") },
		func() string {
			var h, err = os.Hostname()
			xerror.Exit(err)
			return h
		},
	)

	// Namespace 命名空间
	Namespace = utils.FirstNotEmpty(
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

	// Homedir the home directory for the current user
	Homedir = utils.FirstNotEmpty(
		homedir.HomeDir,
		func() string {
			var h, err = dir.Dir()
			xerror.Exit(err)
			return h
		},
		func() string { return "." },
	)
)

func Name() string { return envs.Name() }
