package result

import "github.com/pubgo/funk/xtry"

func ToList[T any](results <-chan Result[T]) List[T] {
	var rr []Result[T]
	for r := range results {
		rr = append(rr, r)
	}
	return rr
}

func CreateList[T any](fn func(in chan T) error) Result[[]T] {
	var data = make(chan T)
	if err := xtry.Try(func() error { return fn(data) }); err != nil {
		return Err[[]T](err)
	}
	close(data)

	var dd []T
	for v := range data {
		dd = append(dd, v)
	}
	return OK(dd)
}

func init() {
	CreateList(func(in chan string) error {
		in <- ""
		in <- ""
		in <- ""
		return nil
	})
}
