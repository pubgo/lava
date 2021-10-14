package syncx

import (
	"context"
	"github.com/pubgo/xerror"
	"sync"
)

func newValue() *AsyncValue { return &AsyncValue{ch: make(chan struct{})} }

type AsyncValue struct {
	done   sync.Once
	value  interface{}
	err    error
	ch     chan struct{}
	cancel context.CancelFunc
}

func (v *AsyncValue) Expect(format string, args ...interface{}) interface{} {
	if v.Err() == nil {
		return v.value
	}

	xerror.PanicF(v.err, format, args...)
	return nil
}

func (v *AsyncValue) ValueCb(f func(val interface{}) error) error {
	if v.Err() == nil {
		return f(v.Get())
	}

	return v.Err()
}

func (v *AsyncValue) Value(f func(err error)) interface{} {
	if v.Err() != nil {
		f(v.Err())
		return nil
	}
	return v.Get()
}

func (v *AsyncValue) Error() string {
	if v.err == nil {
		return ""
	}

	return v.err.Error()
}

func (v *AsyncValue) getVal() interface{} {
	v.done.Do(func() { <-v.ch })
	return v.value
}

func (v *AsyncValue) Cancel()          { v.cancel() }
func (v *AsyncValue) Err() error       { _ = v.getVal(); return v.err }
func (v *AsyncValue) Get() interface{} { return v.getVal() }
