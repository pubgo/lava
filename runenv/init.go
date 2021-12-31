package runenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/denisbrodbeck/machineid"
	dir "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/util/homedir"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/xerror"
)

// 默认的全局配置
var (
	Domain       = version.Domain
	CatchSigpipe = false
	Block        = true
	Trace        = false
	Addr         = ":8080"
	DebugAddr    = ":8081"
	Project      = "lava"
	Level        = "debug"
	Mode         = "dev"

	// DeviceId 设备ID
	DeviceId = xerror.PanicErr(machineid.ID())

	Signal os.Signal = syscall.Signal(0)

	// Pwd 当前目录
	Pwd = xerror.PanicStr(os.Getwd())

	// Hostname 主机名
	Hostname = lavax.FirstNotEmpty(
		func() string { return os.Getenv("HOSTNAME") },
		func() string {
			var h, err = os.Hostname()
			xerror.Exit(err)
			return h
		},
	)

	// Namespace 命名空间
	Namespace = lavax.FirstNotEmpty(
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
		func() string { return "default" },
	)

	// Homedir the home directory for the current user
	Homedir = lavax.FirstNotEmpty(
		homedir.HomeDir,
		func() string {
			var h, err = dir.Dir()
			xerror.Exit(err)
			return h
		},
	)
)

func Name() string {
	return fmt.Sprintf("%s-%s", Domain, Project)
}
