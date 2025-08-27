package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/v2/result"
)

const Name = "pidfile"

var PidPath = filepath.Join(config.GetConfigDir(), Name)

const pidPerm os.FileMode = 0o644

func Get() (r result.Result[int]) {
	pidPath := GetPath().UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	p, err := os.ReadFile(pidPath)
	if err != nil {
		return r.WithErrorf("failed to read pid file: %s", pidPath)
	}

	if len(p) == 0 {
		return r.WithErrorf("pid file is empty")
	}

	return result.Wrap(strconv.Atoi(string(p))).
		InspectErr(func(err error) {
			log.Err(err).Str("path", pidPath).Str("pid", string(p)).Msg("read pid file failed")
		})
}

func GetPath() (r result.Result[string]) {
	filename := fmt.Sprintf("%s.pid", running.Project)
	pidPath := filepath.Join(PidPath, filename)

	if pathutil.IsNotExist(PidPath) {
		createDirRes := result.ErrOf(os.MkdirAll(PidPath, os.ModePerm)).InspectErr(func(err error) {
			log.Err(err).Str("dir", PidPath).Msg("create pid file dir failed")
		})
		if createDirRes.CatchErr(&r) {
			return
		}
	}

	return r.WithValue(pidPath)
}

func Save() (r result.Error) {
	pidPath := GetPath().UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	pid := syscall.Getpid()

	return result.ErrOf(os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), pidPerm)).
		InspectErr(func(err error) {
			log.Err(err).Str("path", pidPath).Int("pid", pid).Msg("write pid file failed")
		})
}
