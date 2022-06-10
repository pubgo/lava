package typex

import (
	"reflect"
)

// StrOf string slice
func StrOf(s1 string, ss ...string) []string {
	return append(append(make([]string, 0, len(ss)+1), s1), ss...)
}

// ObjOf object slice
func ObjOf(s1 interface{}, ss ...interface{}) []interface{} {
	return append(append(make([]interface{}, 0, len(ss)+1), s1), ss...)
}

func ValueOf(s1 reflect.Value, ss ...reflect.Value) []reflect.Value {
	return append(append(make([]reflect.Value, 0, len(ss)+1), s1), ss...)
}
