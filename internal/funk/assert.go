package funk

import "fmt"

func Assert(b bool, format string, a ...interface{}) {
	if b {
		panic(fmt.Errorf(format, a...))
	}
}

func AssertErr(b bool, err error) {
	if b {
		panic(err)
	}
}

func AssertFn(b bool, fn func() error) {
	if b {
		panic(fn())
	}
}
