package connector

import (
	"sync"

	"github.com/pubgo/lug/abc/factory"
	"github.com/pubgo/lug/types"

	"github.com/google/uuid"
	"github.com/pubgo/xerror"
)

var connectors sync.Map

func Get(name string) Connector {
	var val, ok = connectors.Load(name)
	if ok {
		return val.(Connector)
	}

	return nil
}

func List() map[string]Connector {
	var ss = make(map[string]Connector)
	connectors.Range(func(key, value interface{}) bool {
		ss[key.(string)] = value.(Connector)
		return true
	})
	return ss
}

func Init(name string, cfg types.CfgMap) (_ Connector, err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "", "name is null")

	var fct = factory.Get(Name, name)
	xerror.Assert(fct == nil, "connector factory [%s] not found", name)
	xerror.Panic(fct.Init(cfg))

	// 保存connector
	connectors.Store(uuid.New().String(), fct.Get())
	return fct.Get().(Connector), nil
}
