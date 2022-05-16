package main

import (
	"fmt"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
)

type Reg interface {
	Hello()
}

type helloImpl struct {
	ss string
}

func (t *helloImpl) Hello() {

}

func main() {
	xerror.Exit(fx.New(
		fx.Provide(fx.Annotated{
			Group: "jj",
			Target: func() Reg {
				return &helloImpl{ss: "dd"}
			},
		}),
		fx.Provide(fx.Annotated{
			Group: "jj",
			Target: func() Reg {
				return nil
			},
		}), fx.Invoke(func(jj struct {
			fx.In
			Jj []Reg `group:"jj"`
		}) {
			fmt.Println(jj.Jj, len(jj.Jj))
		})).Err())
}
