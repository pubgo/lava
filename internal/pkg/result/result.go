package result

type Result[T any] struct {
	V T
	E error
}

func (v Result[T]) Err() error {
	return v.E
}

func (v Result[T]) IsErr() bool {
	return v.E == nil
}

func (v Result[T]) Get() T {
	return v.V
}
