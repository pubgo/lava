// +build trace

package functrace

import (
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/pubgo/x/stack"
)

var mu sync.Mutex
var m = make(map[uint]int)

func printTrace(id uint, name, typ string, indent int) {
	indents := ""
	for i := 0; i < indent; i++ {
		indents += "\t"
	}
	fmt.Printf("g[%02d]:%s%s%s\n", id, indents, typ, name)
}

func Trace() func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	id, err := stack.GoroutineID()
	if err != nil {
		if err == io.EOF {
			return func() {}
		}
		panic(err)
	}

	fn := runtime.FuncForPC(pc)
	name := fn.Name()

	mu.Lock()
	v := m[id]
	m[id] = v + 1
	mu.Unlock()
	printTrace(id, name, "->", v+1)
	return func() {
		mu.Lock()
		v := m[id]
		m[id] = v - 1
		mu.Unlock()
		printTrace(id, name, "<-", v)
	}
}
