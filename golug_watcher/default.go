package golug_watcher

import (
	"sync"

	"github.com/pubgo/xerror"
)

var dataCallback sync.Map

func Watch(name string, h CallBack) {
	if h == nil {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s is nil", name))
	}

	if _, ok := dataCallback.LoadOrStore(name, h); ok {
		xerror.Next().Panic(xerror.Fmt("[watcher] %s already exists", name))
	}
}

func GetCallBack(name string) CallBack {
	val, ok := dataCallback.Load(name)
	if ok {
		return val.(CallBack)
	}
	return func(event *Response) error { return nil }
}
