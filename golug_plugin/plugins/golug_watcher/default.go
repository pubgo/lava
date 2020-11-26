package golug_watcher

import (
	"errors"

	"github.com/pubgo/xerror"
)

var defaultWatcher = func() *watcherManager {
	w, err := New()
	xerror.Exit(err)
	return w
}()

func SetDefault(c *watcherManager) {
	defaultWatcher = c
}

func getDefault() *watcherManager {
	if defaultWatcher != nil {
		return defaultWatcher
	}

	xerror.Exit(errors.New("please init watcherManager"))
	return nil
}

func Start() {
	getDefault().Start()
}

func Stop() {
	getDefault().Stop()
}

func Add(name string, h CallBack) error {
	return xerror.Wrap(getDefault().Add(name, h))
}

func AddRecursive(name string, h CallBack) (err error) {
	return xerror.Wrap(getDefault().AddRecursive(name, h))
}

func AddExclude(name string) {
	getDefault().AddExclude(name)
}

func Remove(name string) error {
	return xerror.Wrap(getDefault().Remove(name))
}

func List() []string {
	return getDefault().List()
}
