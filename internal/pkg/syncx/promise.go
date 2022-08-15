package syncx

import (
	"github.com/pubgo/lava/internal/pkg/result"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func FromCh[T any](ch chan T) chan result.Result[T] {
	var retCh = make(chan result.Result[T])
	go func() {
		defer close(retCh)
		for cc := range ch {
			retCh <- result.OK(cc)
		}
	}()
	return retCh
}

func From[T any](gen func(in chan<- result.Result[T])) chan result.Future[T] {
	checkFn(gen, "[GoChan] [fn] is nil")

	var ch = make(chan result.Result[T])

	go func() {
		defer close(ch)
		defer xerror.Recovery(func(err xerror.XErr) {
			ch <- result.Err[T](err.Wrap("GoChan", stack.Func(gen)))
		})

		gen(ch)
	}()

	return ch
}
