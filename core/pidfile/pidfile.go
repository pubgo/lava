package pidfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/funk/result"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/runmode"
)

const Name = "pidfile"

var pidPath = filepath.Join(config.CfgDir, "pidfile")

const pidPerm os.FileMode = 0666

func GetPid() result.Result[int] {
	f := GetPidF()
	if f.IsErr() {
		return result.Err[int](f.Err())
	}

	p, err := ioutil.ReadFile(f.Unwrap())
	if err != nil {
		return result.Wrap(0, err)
	}

	return result.Wrap(strconv.Atoi(string(p)))
}

func GetPidF() result.Result[string] {
	filename := fmt.Sprintf("%s.pid", runmode.Project)
	return result.OK(filepath.Join(pidPath, filename))
}

func SavePid() error {
	f := GetPidF()
	if f.IsErr() {
		return f.Err()
	}

	pid := syscall.Getpid()
	return ioutil.WriteFile(f.Unwrap(), []byte(strconv.Itoa(pid)), pidPerm)
}
