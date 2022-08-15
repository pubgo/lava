package result

func NewFuture[T any]() Future[T] {
	return Future[T]{done: make(chan struct{})}
}

type Future[T any] struct {
	V    T
	E    error
	done chan struct{}
}

func (v Future[T]) Await() Result[T] {
	<-v.done
	return Result[T]{V: v.V, E: v.E}
}

func (v Future[T]) Done() {
	close(v.done)
}
