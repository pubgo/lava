package types

import (
	"reflect"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
)

func NewSMap() *SMap { return &SMap{} }

type SMap struct {
	data  sync.Map
	count atomic.Uint32
}

func (t *SMap) Each(fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	vfn := reflect.ValueOf(fn)
	isKey := vfn.Type().NumIn() == 1

	t.data.Range(func(key, value interface{}) bool {
		if isKey {
			vfn.Call([]reflect.Value{reflect.ValueOf(key)})
			return true
		}

		vfn.Call([]reflect.Value{reflect.ValueOf(key), reflect.ValueOf(value)})
		return true
	})

	return nil
}

func (t *SMap) Map(data interface{}) (err error) {
	defer xerror.RespErr(&err)

	vd := reflect.ValueOf(data)
	if vd.Kind() == reflect.Ptr {
		vd = vd.Elem()
		vd.Set(reflect.MakeMap(vd.Type()))
	}

	// var data = make(map[string]int); Map(data)
	// var data map[string]int; Map(&data)
	xerror.Assert(!vd.IsValid() || vd.IsNil(), "[data] type error")

	t.data.Range(func(key, value interface{}) bool {
		vd.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return true
	})

	return nil
}

func (t *SMap) Set(key, value interface{}) {
	_, ok := t.data.LoadOrStore(key, value)
	if !ok {
		t.count.Inc()
	}
}

func (t *SMap) Load(key interface{}) (value interface{}, ok bool) { return t.data.Load(key) }
func (t *SMap) Range(f func(key, value interface{}) bool)         { t.data.Range(f) }
func (t *SMap) Len() int                                          { return int(t.count.Load()) }
func (t *SMap) Delete(key interface{})                            { t.data.Delete(key); t.count.Dec() }
func (t *SMap) Get(key interface{}) (value interface{})           { value, _ = t.data.Load(key); return }
func (t *SMap) Has(key interface{}) (ok bool)                     { _, ok = t.data.Load(key); return }
