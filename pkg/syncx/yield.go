package syncx

import "github.com/pubgo/xerror"

func Wait(val ...chan Value) []Value {
	var valList = make([]Value, len(val))
	for i := range val {
		valList[i] = <-val[i]
	}
	return valList
}

// Async 通过chan的方式同步执行异步任务
func Async(fn func() Value) chan Value { return GoChan(fn) }

func newPromise() *Promise { return &Promise{ch: make(chan interface{})} }

type Promise struct {
	err error
	ch  chan interface{}
}

func (t *Promise) Err() error                  { return t.err }
func (t *Promise) close()                      { close(t.ch) }
func (t *Promise) Unwrap() <-chan interface{}  { return t.ch }
func (t *Promise) Await() (interface{}, error) { return <-t.ch, t.err }
func (t *Promise) Range(fn func(interface{})) error {
	if t.err != nil {
		return t.err
	}

	for v := range t.ch {
		fn(v)
	}

	return t.err
}

func Yield(fn func() (interface{}, error)) *Promise {
	if fn == nil {
		panic("[Yield] [fn] should not be nil")
	}

	var p = newPromise()
	go func() {
		defer func() {
			p.close()
			xerror.RespErr(&p.err)
		}()

		val, err := fn()
		p.err = err
		p.ch <- val
	}()

	return p
}

func YieldGroup(fn func(in chan<- *Promise) error) *Promise {
	if fn == nil {
		panic("[YieldGroup] [fn] should not be nil")
	}

	var p = &Promise{ch: make(chan interface{})}
	var in = make(chan *Promise)

	go func() {
		defer p.close()
		for pp := range in {
			for val := range pp.Unwrap() {
				p.ch <- val
			}

			if pp.Err() != nil {
				p.err = pp.Err()
			}
		}
	}()

	go func() {
		defer close(in)
		defer xerror.RespErr(&p.err)
		if err := fn(in); err != nil {
			p.err = err
		}
	}()

	return p
}
