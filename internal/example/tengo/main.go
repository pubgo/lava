package main

import (
	"fmt"
	"github.com/d5/tengo/v2/token"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
)

func main() {
	s := tengo.NewScript([]byte(`
	
	b := a > "hello"
	a = a | "world"

`))

	// define variable 'b'
	_ = s.Add("a", &obj{})

	s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	// compile the source
	c, err := s.Compile()
	if err != nil {
		panic(err)
	}

	// run the compiled bytecode
	// a compiled bytecode 'c' can be executed multiple times without re-compiling it
	if err := c.Run(); err != nil {
		panic(err)
	}

	// retrieve value of 'a'
	fmt.Println(c.Get("a"))
	fmt.Println(c.Get("b"))
}

var _ tengo.Object = (*obj)(nil)

type obj struct {
	data []interface{}
}

func (o *obj) TypeName() string {
	return "obj"
}

func (o *obj) String() string {
	return fmt.Sprint(o.data...)
}

func (o *obj) BinaryOp(op token.Token, rhs tengo.Object) (tengo.Object, error) {
	o.data = append(o.data, op.String(), rhs.String())
	return o, nil
}

func (o *obj) IsFalsy() bool {
	return true
}

func (o *obj) Equals(another tengo.Object) bool {
	panic("implement me")
}

func (o *obj) Copy() tengo.Object {
	panic("implement me")
}

func (o *obj) IndexGet(index tengo.Object) (value tengo.Object, err error) {
	panic("implement me")
}

func (o *obj) IndexSet(index, value tengo.Object) error {
	panic("implement me")
}

func (o *obj) Iterate() tengo.Iterator {
	panic("implement me")
}

func (o *obj) CanIterate() bool {
	panic("implement me")
}

func (o *obj) Call(args ...tengo.Object) (ret tengo.Object, err error) {
	panic("implement me")
}

func (o *obj) CanCall() bool {
	panic("implement me")
}
