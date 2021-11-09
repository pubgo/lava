package tengo

import (
	"fmt"

	tengo "github.com/d5/tengo/v2"
)

func init() {
	s := tengo.NewScript([]byte(`a := b + 20`))

	// define variable 'b'
	_ = s.Add("b", 10)

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
	a := c.Get("a")
	fmt.Println(a.Int()) // prints "30"

	// re-run after replacing value of 'b'
	if err := c.Set("b", 20); err != nil {
		panic(err)
	}
	if err := c.Run(); err != nil {
		panic(err)
	}
	fmt.Println(c.Get("a").Int()) // prints "40"
}

type vm struct {
}
