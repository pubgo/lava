package golug_watcher

import (
	"errors"
	"sync"

	"github.com/pubgo/xerror"
)

var watchers []Watcher
var data sync.Map
var mu sync.Mutex

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

func AddWatcher(c Watcher) {
	mu.Lock()
	defer mu.Unlock()
	watchers = append(watchers, c)
}

func getDefault() []Watcher {
	if len(watchers) != 0 {
		return watchers
	}

	xerror.Panic(errors.New("please init Watcher"))
	return nil
}

func Watch(name string, h CallBack) {
	if h == nil {
		panic(xerror.New("[CallBack] is nil"))
	}

	data.Store(name, h)
}

func GetCallBack(name string) CallBack {
	val, ok := data.Load(name)
	if ok {
		return val.(CallBack)
	}
	return nil
}

func Remove(name string) {
	data.Delete(name)
}

func List() []string {
	var dt []string
	data.Range(func(key, _ interface{}) bool { dt = append(dt, key.(string)); return true })
	return dt
}
