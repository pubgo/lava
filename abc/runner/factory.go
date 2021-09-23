package runner

import (
	"sync"

	"github.com/pubgo/xerror"
)

var runners sync.Map

// Register 注册runner
func Register(id string, runner Runner) {
	xerror.Assert(id == "" || runner == nil, "id==null or runner==nil")

	_, loaded := runners.LoadOrStore(id, runner)
	xerror.Assert(loaded, "[runner] %s already exists", id)
}

// Get get runner
func Get(id string) Runner {
	value, ok := runners.Load(id)
	if ok {
		return value.(Runner)
	}
	return nil
}

// List 获取runner list
func List() map[string]Runner {
	var mapList = make(map[string]Runner)
	runners.Range(func(k, v interface{}) bool {
		mapList[k.(string)] = v.(Runner)
		return true
	})
	return mapList
}
