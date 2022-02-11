package resource

import (
	"io"
	"sync"
)

// Release 释放资源
type Release interface {
	Release()
}

// Resource 资源对象接口
// 	Resource 是真实对象的wrapper, 可以通过getObj获取真实的内部对象
type Resource interface {
	// GetObj 获取资源内部真实对象
	GetObj() io.Closer
	// updateObj 更新资源对象
	updateObj(obj io.Closer)

	// Kind 资源类型
	Kind() string

	// LoadObj 获取真实的资源对象
	//	r: 用完对象后记得 release, 不release会死锁
	LoadObj() (obj interface{}, r Release)
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

func (t *baseRes) Release() { t.rw.RUnlock() }

func (t *baseRes) LoadObj() (interface{}, Release) {
	t.rw.RLock()
	return t.v, t
}

func (t *baseRes) GetObj() io.Closer {
	t.rw.RLock()
	var v = t.v
	t.rw.RUnlock()
	return v
}

func (t *baseRes) updateObj(obj io.Closer) {
	// TODO 注意：使用方不释放资源，会永远阻塞
	t.rw.Lock()
	t.v = obj
	t.rw.Unlock()
}
