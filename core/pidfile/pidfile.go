package pidfile

import (
	"fmt"
	"github.com/pubgo/lava/core/runmode"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/lava/config"
)

const Name = "pidfile"

var pidPath = filepath.Join(config.CfgDir, "pidfile")

const pidPerm os.FileMode = 0666

func GetPid() (int, error) {
	f, err := GetPidF()
	if err != nil {
		return 0, err
	}

	p, err := ioutil.ReadFile(f)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(p))
}

func GetPidF() (string, error) {
	filename := fmt.Sprintf("%s.pid", runmode.Project)
	return filepath.Join(pidPath, filename), nil
}

func SavePid() error {
	f, err := GetPidF()
	if err != nil {
		return err
	}

	pid := syscall.Getpid()
	return ioutil.WriteFile(f, []byte(strconv.Itoa(pid)), pidPerm)
}
