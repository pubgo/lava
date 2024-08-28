package typex

func DoFunc(fn func()) {
	fn()
}

func DoFunc1[T any](fn func() T) T {
	return fn()
}
