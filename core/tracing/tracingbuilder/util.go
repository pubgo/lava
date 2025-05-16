package tracingbuilder

import (
	"runtime"
)

func queueSize() int {
	const minSize = 1000
	const maxSize = 16000

	n := (runtime.GOMAXPROCS(0) / 2) * 1000
	if n < minSize {
		return minSize
	}
	if n > maxSize {
		return maxSize
	}
	return n
}
