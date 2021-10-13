package async

import (
	"sync"
)

type Value interface {
	Err() error
	Get() interface{}
	Cancel()
	Value(func(err error)) interface{}
	ValueCb(func(val interface{}) error) error
	complete(interface{}, error)
	context() Ctx
}

var _ Value = (*valueImpl)(nil)

func newValue() Value {
	return &valueImpl{valChan: make(chan interface{}), ctx: &ctxImpl{}}
}

type valueImpl struct {
	done    sync.Once
	value   interface{}
	err     error
	valChan chan interface{}
	ctx     Ctx
}

func (v *valueImpl) ValueCb(f func(val interface{}) error) error {
	if v.Err() == nil {
		return f(v.Get())
	}

	return v.Err()
}

func (v *valueImpl) Value(f func(err error)) interface{} {
	if v.Err() != nil {
		f(v.Err())
		return nil
	}

	return v.Get()
}

func (v *valueImpl) complete(i interface{}, err error) {
	v.err = err
	v.valChan <- i
}

func (v *valueImpl) Error() string {
	if v.err == nil {
		return ""
	}

	return v.err.Error()
}

func (v *valueImpl) getVal() interface{} {
	v.done.Do(func() {
		if v.valChan != nil {
			defer close(v.valChan)
			v.value = <-v.valChan
		}
	})
	return v.value
}

func (v *valueImpl) context() Ctx     { return v.ctx }
func (v *valueImpl) Cancel()          { v.ctx.Cancel() }
func (v *valueImpl) Err() error       { _ = v.getVal(); return v.err }
func (v *valueImpl) Get() interface{} { return v.getVal() }
