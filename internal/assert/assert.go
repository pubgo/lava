package assert

import (
	"fmt"
)

func Func(ok bool, fn func() error) {
	if ok {
		panic(fn())
	}
}

func Err(ok bool, err error) {
	if ok {
		panic(err)
	}
}

func Msg(ok bool, msg string, args ...interface{}) {
	if ok {
		panic(fmt.Errorf(msg, args...))
	}
}
