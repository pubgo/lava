// Package pidfile 管理进程 PID 文件的读写，用于进程存活检测与单实例守护。
//
// PID 文件默认保存在配置目录下，文件名为 ".{project}.pid"。
// 通过 pidfilebuilder 在服务启动后写入、停止后删除。
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/running"
)

var getPidPath = sync.OnceValue(func() string {
	return filepath.Join(config.GetConfigDir(), "."+running.Project()+".pid")
})

const pidPerm os.FileMode = 0o644

// GetPath 返回 PID 文件的绝对路径。
func GetPath() string { return getPidPath() }

// Get 读取 PID 文件并返回进程 ID，文件不存在或内容非法时返回错误。
func Get() (r result.Result[int]) {
	return readPID(GetPath())
}

// Save 将当前进程 ID 写入 PID 文件。
func Save() (r result.Error) {
	return writePID(GetPath(), os.Getpid())
}

// Remove 删除 PID 文件，通常在进程优雅退出时调用。文件不存在时不视为错误。
func Remove() (r result.Error) {
	pidPath := GetPath()
	err := os.Remove(pidPath)
	if err != nil && !os.IsNotExist(err) {
		return result.ErrOf(err).Log(func(e result.Event) {
			e.Str("path", pidPath)
			e.Msg("remove pid file failed")
		})
	}
	return result.ErrOf(nil)
}

func readPID(pidPath string) (r result.Result[int]) {
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

func writePID(pidPath string, pid int) (r result.Error) {
	return result.ErrOf(os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), pidPerm)).
		Log(func(e result.Event) {
			e.Str("path", pidPath)
			e.Int("pid", pid)
			e.Msg("write pid file failed")
		})
}
