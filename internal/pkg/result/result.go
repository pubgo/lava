package result

func Unwrap[T any](r Result[T]) (T, error) {
	return r.v, r.e
}

func OK[T any](v T) Result[T] {
	return Result[T]{v: v}
}

func Err[T any](err error) Result[T] {
	return Result[T]{e: err}
}

func New[T any](v T, err error) Result[T] {
	return Result[T]{v: v, e: err}
}

type Result[T any] struct {
	v T
	e error
}

func (v Result[T]) WithErr(err error) Result[T] {
	v.e = err
	return v
}

func (v Result[T]) WithVal(t T) Result[T] {
	v.v = t
	return v
}

func (v Result[T]) Err() error {
	return v.e
}

func (v Result[T]) IsErr() bool {
	return v.e != nil
}

func (v Result[T]) Get() T {
	return v.v
}

type Chan[T any] chan Result[T]

func (cc Chan[T]) ToList() List[T] {
	return ToList(cc)
}

func (cc Chan[T]) Range(fn func(r Result[T])) {
	for c := range cc {
		fn(c)
	}
}

type List[T any] []Result[T]

func (rr List[T]) ToResult() Result[[]T] {
	var rl = make([]T, 0, len(rr))
	for i := range rr {
		if rr[i].IsErr() {
			return Err[[]T](rr[i].Err())
		}
		rl = append(rl, rr[i].Get())
	}
	return OK(rl)
}

func (rr List[T]) Range(fn func(r Result[T])) {
	for i := range rr {
		fn(rr[i])
	}
}
