package result

func ToList[T any](results <-chan Result[T]) List[T] {
	var rr []Result[T]
	for r := range results {
		rr = append(rr, r)
	}
	return rr
}
