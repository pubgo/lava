package syncx

import (
	"sync"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/stack"

	"github.com/pubgo/lava/internal/pkg/result"
)

func NewFuture[T any]() *Future[T] {
	return &Future[T]{done: make(chan struct{})}
}

type Future[T any] struct {
	v    T
	e    error
	done chan struct{}
	once sync.Once
}

func (f Future[T]) OK(v T) {
	f.once.Do(func() {
		f.v = v
		close(f.done)
	})
}

func (f Future[T]) Err(err error) {
	f.once.Do(func() {
		f.e = err
		close(f.done)
	})
}

func (f Future[T]) Await() result.Result[T] {
	<-f.done
	return result.New(f.v, f.e)
}

func FromFuture[T any](gen func(in chan<- *Future[T])) result.Chan[T] {
	checkFn(gen, "[GoChan] [fn] is nil")

	var ch = make(chan *Future[T])
	var retCh = make(chan result.Result[T])
	go func() {
		defer close(retCh)
		for cc := range ch {
			retCh <- cc.Await()
		}
	}()

	go func() {
		defer close(ch)
		defer recovery.Recovery(func(err xerr.XErr) {
			var f = NewFuture[T]()
			f.Err(err.Wrap("GoChan", stack.Func(gen)))
			ch <- f
		})
		gen(ch)
	}()

	return retCh
}

func FromCh[T any](ch chan T) result.Chan[T] {
	var retCh = make(chan result.Result[T])
	go func() {
		defer close(retCh)
		for cc := range ch {
			retCh <- result.OK(cc)
		}
	}()
	return retCh
}

func From[T any](gen func(in chan<- result.Result[T])) result.Chan[T] {
	checkFn(gen, "[GoChan] [fn] is nil")

	var ch = make(chan result.Result[T])

	go func() {
		defer close(ch)
		defer recovery.Recovery(func(err xerr.XErr) {
			ch <- result.Err[T](err.Wrap("GoChan", stack.Func(gen)))
		})

		gen(ch)
	}()

	return ch
}

func Wait[T any](val ...*Future[T]) result.List[T] {
	var valList = make([]result.Result[T], len(val))
	for i := range val {
		valList[i] = val[i].Await()
	}
	return valList
}

// Async 通过chan的方式同步执行异步任务
func Async[T any](fn func() result.Result[T]) *Future[T] { return GoChan(fn) }
