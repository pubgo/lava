package typex

import "testing"

func BenchmarkStrOf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StrOf("hello", "hello", "hello", "hello")
	}
}
