package reflectx

import "reflect"

func Indirect(v reflect.Value) reflect.Value {
	for {
		if v.Kind() != reflect.Ptr {
			return v
		}
		v = v.Elem()
	}
}
