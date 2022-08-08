package typex

//func Result[T any]() (T, error) {
//
//}

func Err[T any](err error) Value[T] {
	return Value[T]{err: err}
}

func OK[T any](v T, errs ...error) Value[T] {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	return Value[T]{v: v, err: err}
}

type Value[T any] struct {
	v   T
	err error
}

func (v Value[T]) Err() error {
	return v.err
}

func (v Value[T]) IsErr() bool {
	return v.err == nil
}

func (v Value[T]) Get() T {
	return v.v
}
