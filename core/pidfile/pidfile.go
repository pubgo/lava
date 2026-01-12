package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/running"
)

var getPidPath = sync.OnceValue(func() string {
	return filepath.Join(config.GetConfigDir(), "."+running.Project()+".pid")
})

const pidPerm os.FileMode = 0o644

func Get() (r result.Result[int]) {
	pidPath := GetPath()
	p := result.Wrap(os.ReadFile(pidPath)).
		Validate(func(val []byte) error {
			if len(val) == 0 {
				return fmt.Errorf("pid file is empty")
			}
			return nil
		}).
		Log(func(e result.Event) {
			e.Str("path", pidPath)
			e.Msg("read pid file failed")
		}).
		UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	return result.Wrap(strconv.Atoi(string(p))).
		Log(func(e result.Event) {
			e.Str("path", pidPath)
			e.Str("pid", string(p))
			e.Msg("convert pid to int failed")
		})
}

func GetPath() string { return getPidPath() }

func Save() (r result.Error) {
	pidPath := GetPath()
	pid := os.Getpid()

	return result.ErrOf(os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), pidPerm)).
		Log(func(e result.Event) {
			e.Str("path", pidPath)
			e.Int("pid", pid)
			e.Msg("write pid file failed")
		})
}
