package golug_watcher

import (
	"errors"

	"github.com/pubgo/xerror"
)

var watchers []Watcher

func AddWatcher(c Watcher) {
	watchers = append(watchers, c)
}

func getDefault() []Watcher {
	if len(watchers) != 0 {
		return watchers
	}

	xerror.Exit(errors.New("please init Watcher"))
	return nil
}

func Start() {
	for _, w := range getDefault() {
		xerror.ExitF(w.Start(), w.String())
	}
}

func Close() {
	for _, w := range getDefault() {
		xerror.ExitF(w.Close(), w.String())
	}
}

func Watch(name string, h CallBack) {
	for _, w := range getDefault() {
		xerror.ExitF(w.Watch(name, h), "name:%s watcher:%s", name, w.String())
	}
}

func Remove(name string) {
	for _, w := range getDefault() {
		xerror.ExitF(w.Remove(name), "name:%s watcher:%s", name, w.String())
	}
}

func List() []string {
	return getDefault()[0].List()
}
