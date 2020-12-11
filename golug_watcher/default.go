package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var dataCallback sync.Map

func Watch(name string, h CallBack) {
	if h == nil {
		xerror.Next().Panic(xerror.New("[CallBack] is nil"))
	}

	dataCallback.Store(name, h)
}

func GetCallBack(name string) CallBack {
	val, ok := dataCallback.Load(name)
	if ok {
		return val.(CallBack)
	}
	return nil
}
