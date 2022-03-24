package main

import (
	"fmt"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/inject"
	"reflect"
)

type IHello interface {
}

var _ error = (*hello)(nil)

type hello struct {
	IHello `json:"i_hello"`
	L      *logging.Logger
	Name   string
}

func (h hello) Error() string {
	h.L.Info("hello test")
	return ""
}

func Register(err error) {
	inject.Inject(err)
	err.Error()
}

func init1() IHello {
	return nil
}

func main() {
	var v = reflect.TypeOf((*IHello)(nil))
	fmt.Println(v.Elem().Name())

	//fmt.Println(reflect.TypeOf(&bbolt.Client{}).String())
	Register(&hello{})
	inject.Register((*IHello)(nil), func(obj inject.Object, field inject.Field) (interface{}, bool) {
		return nil, true
	})
}
