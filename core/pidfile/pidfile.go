package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/running"
)

const Name = "pidfile"

var PidPath = filepath.Join(config.GetConfigDir(), Name)

const pidPerm os.FileMode = 0o666

func GetPid() result.Result[int] {
	f := GetPidF()
	if f.IsErr() {
		return result.Err[int](f.Err())
	}

	p, err := os.ReadFile(f.Unwrap())
	if err != nil {
		return result.Wrap(0, err)
	}

	return result.Wrap(strconv.Atoi(string(p)))
}

func GetPidF() result.Result[string] {
	filename := fmt.Sprintf("%s.pid", running.Project)
	return result.OK(filepath.Join(PidPath, filename))
}

func SavePid() error {
	f := GetPidF()
	if f.IsErr() {
		return f.Err()
	}

	pid := syscall.Getpid()
	return os.WriteFile(f.Unwrap(), []byte(strconv.Itoa(pid)), pidPerm)
}
