package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var data sync.Map

func Start() {
	for _, w := range List() {
		xerror.ExitF(w.Start(), w.String())
	}
}

func Close() {
	for _, w := range List() {
		xerror.ExitF(w.Close(), w.String())
	}
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
