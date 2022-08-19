package reflectx

import (
	"reflect"
)

func Indirect(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		panic("[v] is invalid")
	}

	for {
		if v.Kind() != reflect.Ptr {
			return v
		}
		v = v.Elem()
	}
}

func New(val interface{}) reflect.Value {
	if val == nil {
		panic("[val] is nil")
	}

	return reflect.New(Indirect(reflect.ValueOf(val)).Type())
}

func FindFieldBy(v reflect.Value, handle func(field reflect.StructField) bool) reflect.Value {
	var t = v.Type()
	for i := v.NumField() - 1; i >= 0; i-- {
		if handle(t.Field(i)) {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}
