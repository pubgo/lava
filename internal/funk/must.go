package funk

func Must(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func Must1[R any](r R, err error) R {
	if err == nil {
		return r
	}
	panic(err)
}
