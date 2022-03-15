package main

import (
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging"
	"go.uber.org/zap"
)

type IHello interface {
}

var _ error = (*hello)(nil)

type hello struct {
	IHello `json:"i_hello"`
	L      *zap.Logger
	Name   string
}

func (h hello) Error() string {
	h.L.Info("hello test")
	return ""
}

func Register(err error) {
	_ = logging.L()
	inject.Inject(err)
	err.Error()
}

func main() {
	//fmt.Println(reflect.TypeOf(&bbolt.Client{}).String())
	Register(&hello{})
}
