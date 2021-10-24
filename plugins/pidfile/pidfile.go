package pidfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pubgo/x/pathutil"
)

var pidPath = filepath.Join(os.Getenv("HOME"), "pidfile")

const pidPerm os.FileMode = 0666

type PidManager struct {
	name string
}

func New(name string) *PidManager {
	return &PidManager{name: name}
}

func (pm *PidManager) GetPid() (int, error) {
	f, err := pm.GetPidF()
	if err != nil {
		return 0, err
	}

	p, err := ioutil.ReadFile(f)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(p))
}

func (pm *PidManager) GetPidF() (string, error) {
	if !pathutil.Exist(pidPath) {
		if err := os.MkdirAll(pidPath, pidPerm); err != nil {
			return "", err
		}
	}

	filename := fmt.Sprintf("%s-%s.pid", pm.GetBinName(), pm.name)
	fullPath := filepath.Join(pidPath, filename)
	return fullPath, nil
}

func (pm *PidManager) SavePid() error {
	f, err := pm.GetPidF()
	if err != nil {
		return err
	}

	pid := syscall.Getpid()
	return ioutil.WriteFile(f, []byte(strconv.Itoa(pid)), pidPerm)
}

func (pm *PidManager) GetBinName() string {
	return filepath.Base(os.Args[0])
}
