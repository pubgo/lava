package typex

import (
	"reflect"
	"sync"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
)

func SetOf(val ...interface{}) *Set {
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

func (t *Set) Has(v interface{}) bool { _, ok := t.m.Load(v); return ok }
func (t *Set) Len() uint32            { return t.count.Load() }

func (t *Set) Map(data interface{}) (err error) {
	defer xerror.RespErr(&err)

	vd := reflect.ValueOf(data)
	xerror.Assert(vd.Kind() != reflect.Ptr, "[data] should be ptr type")
	vd = vd.Elem()

	dt := reflect.MakeSlice(vd.Type(), 0, int(t.count.Load()))
	t.m.Range(func(key, _ interface{}) bool {
		dt = reflect.AppendSlice(dt, reflect.ValueOf(key))
		return true
	})
	vd.Set(dt)

	return nil
}

func (t *Set) Add(v interface{}) {
	_, ok := t.m.LoadOrStore(v, struct{}{})
	if !ok {
		t.count.Inc()
	}
}

func (t *Set) List() (val []interface{}) {
	t.m.Range(func(key, _ interface{}) bool { val = append(val, key); return true })
	return
}

func (t *Set) Each(fn interface{}) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	vfn := fx.WrapRaw(fn)
	t.m.Range(func(key, value interface{}) bool { _ = vfn(key); return true })
}
