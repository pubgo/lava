package main

import (
	"fmt"
	"reflect"
)

type hello struct {
	Data *string
}

type hello1 struct {
	Hello hello
	h     hello
	H1    *hello
}

func main() {
	var h hello1
	var v = reflect.ValueOf(&h)
	fmt.Println(v.Kind())
	fmt.Println(v.CanSet())
	fmt.Println(v.Elem().CanSet())
	v = v.Elem()
	for i := 0; i < v.NumField(); i++ {
		fmt.Println(v.Field(i).CanSet())
		//v1 := v.Field(i)
		//for j := 0; j < v1.NumField(); j++ {
		//	fmt.Println(v1.Field(i).CanSet())
		//	var s = "hello"
		//	v1.Field(i).Set(reflect.ValueOf(&s))
		//}
	}
	fmt.Println(h)
}
