package golug_types

import (
	"go.uber.org/atomic"
	"sync"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xerror/xerror_util"
)

func NewSet(val ...interface{}) *Set {
	s := &Set{}
	for i := range val {
		s.Add(val[i])
	}
	return s
}

type Set struct {
	m     sync.Map
	count atomic.Uint32
}

func (t *Set) Len() uint32 { return t.count.Load() }
func (t *Set) Add(v interface{}) {
	_, ok := t.m.LoadOrStore(v, struct{}{})
	if !ok {
		t.count.Inc()
	}
}
func (t *Set) Has(v interface{}) bool { _, ok := t.m.Load(v); return ok }
func (t *Set) List() (val []interface{}) {
	t.m.Range(func(key, _ interface{}) bool { val = append(val, key); return true })
	return
}
func (t *Set) Each(fn interface{}) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	vfn := xerror_util.FuncRaw(fn)
	t.m.Range(func(key, value interface{}) bool { _ = vfn(key); return true })
}
