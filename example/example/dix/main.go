package main

import (
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
)

func main() {
	var i = 10
	xerror.Panic(dix.Provider(&i))

	var m *int
	xerror.Panic(dix.Inject(&m))

	if m != nil {
		fmt.Println(*m)
	}

	fmt.Println(dix.Graph())
}
