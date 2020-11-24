package golug_pidfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
)

const pidPerm os.FileMode = 0755

func GetPid() (pid int, _ error) {
	pidData, err := ioutil.ReadFile(GetPidPath())
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	pid, err = strconv.Atoi(string(pidData))
	return pid, xerror.Wrap(err)
}

func SavePid() error {
	pidBytes := []byte(strconv.Itoa(os.Getpid()))
	return xerror.Wrap(ioutil.WriteFile(GetPidPath(), pidBytes, pidPerm))
}

func GetPidPath() string {
	return filepath.Join(golug_env.Home, "pidfile", golug_env.Domain+"."+golug_env.Project+".pid")
}

func init() {
	// 检查存放pid的目录是否存在, 不存在就创建
	xerror.Panic(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		pidPath := filepath.Dir(GetPidPath())
		if !golug_util.PathExist(pidPath) {
			xerror.Exit(os.MkdirAll(pidPath, pidPerm))
		}
	}))

	// 保存pid到文件当中
	//xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) { xerror.Panic(SavePid()) }))
}
