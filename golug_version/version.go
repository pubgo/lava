package golug_version

import (
	"sync"

	"github.com/imdario/mergo"
	"github.com/pubgo/xerror"
)

var versions sync.Map

type M = map[string]string

func Register(name string, data M) {
	_, ok := versions.LoadOrStore(name, data)
	if ok {
		xerror.Next().Panic(xerror.Fmt("%s already exists", name))
	}
}

func Get(name string) (v M) {
	m, ok := versions.Load(name)
	if ok {
		xerror.Panic(mergo.Map(&v, m))
		return
	}
	return
}

func List() map[string]M {
	ms := make(map[string]M)
	versions.Range(func(key, value interface{}) bool {
		var v M
		xerror.Panic(mergo.Map(&v, value))
		ms[key.(string)] = v
		return true
	})
	return ms
}
