package syncx

type Value interface {
	Val() interface{}
	Err() error
	IsErr() bool
}

var _ Value = (*valueImpl)(nil)

type valueImpl struct {
	err error
	val interface{}
}

func (v *valueImpl) IsErr() bool      { return v.err != nil }
func (v *valueImpl) Val() interface{} { return v.val }
func (v *valueImpl) Err() error       { return v.err }

func WithVal(val interface{}, err ...error) Value {
	var e error
	if len(err) > 0 {
		e = err[0]
	}
	return &valueImpl{val: val, err: e}
}

func WithErr(err error) Value {
	return &valueImpl{err: err}
}
