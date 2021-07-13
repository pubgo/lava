package typex

import (
	"reflect"
	"sync"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
)

var NotFound = new(struct{})

type SMap struct {
	data sync.Map
}

func (t *SMap) Each(fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(fn == nil, "[fn] should not be nil")

	vfn := fx.WrapRaw(fn)
	onlyKey := reflect.TypeOf(fn).NumIn() == 1
	t.data.Range(func(key, value interface{}) bool {
		if onlyKey {
			_ = vfn(key)
			return true
		}

		_ = vfn(key, value)
		return true
	})

	return nil
}

func (t *SMap) Map(fn func(val interface{}) interface{}) {
	t.data.Range(func(key, value interface{}) bool {
		t.data.Store(key, fn(value))
		return true
	})
}

func (t *SMap) MapTo(data interface{}) (err error) {
	defer xerror.RespErr(&err)

	vd := reflect.ValueOf(data)
	if vd.Kind() == reflect.Ptr {
		vd = vd.Elem()
		vd.Set(reflect.MakeMap(vd.Type()))
	}

	// var data = make(map[string]int); MapTo(data)
	// var data map[string]int; MapTo(&data)
	xerror.Assert(!vd.IsValid() || vd.IsNil(), "[data] type error")

	t.data.Range(func(key, value interface{}) bool {
		vd.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return true
	})

	return nil
}

func (t *SMap) Set(key, value interface{}) {
	t.data.Store(key, value)
}

func (t *SMap) Get(key interface{}) interface{} {
	value, ok := t.data.Load(key)
	if ok {
		return value
	}

	return NotFound
}

func (t *SMap) LoadAndDelete(key interface{}) (value interface{}, ok bool) { return t.data.LoadAndDelete(key) }
func (t *SMap) Load(key interface{}) (value interface{}, ok bool) { return t.data.Load(key) }
func (t *SMap) Range(f func(key, value interface{}) bool)         { t.data.Range(f) }
func (t *SMap) Delete(key interface{})                            { t.data.Delete(key) }
func (t *SMap) Has(key interface{}) (ok bool)                     { _, ok = t.data.Load(key); return }
