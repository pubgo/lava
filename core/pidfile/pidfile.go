package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/log/logfields"
	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/running"
	"github.com/rs/zerolog"
)

const Name = "pidfile"

var PidPath = filepath.Join(config.GetConfigDir(), Name)

const pidPerm os.FileMode = 0o644

func Get() (r result.Result[int]) {
	pidPath := GetPath().Unwrap(&r)
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
	filename := fmt.Sprintf("%s.pid", running.Project())
	pidPath := filepath.Join(PidPath, filename)

	if pathutil.IsNotExist(PidPath) {
		createDirRes := result.ErrOf(os.MkdirAll(PidPath, os.ModePerm)).Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, fmt.Sprintf("create pid file dir(%s) failed", PidPath))
		})
		if createDirRes.Catch(&r) {
			return
		}
	}

	return r.WithValue(pidPath)
}

func Save() (r result.Error) {
	pidPath := GetPath().Unwrap(&r)
	if r.IsErr() {
		return
	}

	pid := syscall.Getpid()

	return result.ErrOf(os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), pidPerm)).
		Log(func(e *zerolog.Event) {
			e.Str("path", pidPath)
			e.Int("pid", pid)
			e.Str(logfields.Msg, "write pid file failed")
		})
}
