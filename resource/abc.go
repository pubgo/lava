package resource

import (
	"context"
	"io"
	"sync"
)

// Resource 资源对象接口
// 	Resource 是真实对象的wrapper, 可以通过Unwrap获取真实的内部对象
type Resource interface {
	// getObj 获取资源内部真实对象
	getObj() io.Closer
	// updateObj 获取资源对象
	updateObj(obj io.Closer)

	// Kind 资源类型
	Kind() string

	// Load 获取真实的资源对象
	//	release 用完对象后记得release
	Load() (obj interface{}, cancel context.CancelFunc)
}

func New(val io.Closer) Resource {
	return &baseRes{v: val}
}

type baseRes struct {
	v  io.Closer
	rw sync.RWMutex
}

func (t *baseRes) Kind() string {
	panic("implement me")
}

func (t *baseRes) Load() (interface{}, context.CancelFunc) {
	t.rw.RLock()
	return t.v, t.rw.RUnlock
}

func (t *baseRes) getObj() io.Closer {
	t.rw.RLock()
	var v = t.v
	t.rw.RUnlock()
	return v
}

func (t *baseRes) updateObj(obj io.Closer) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.v = obj
}
