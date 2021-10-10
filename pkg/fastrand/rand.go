package fastrand

import (
	"github.com/valyala/fastrand"
	"testing"
)

func TestName(t *testing.T) {

}

// Uint32 returns pseudorandom uint32.
// It is safe calling this function from concurrent goroutines.
func Uint32() uint32 { return fastrand.Uint32() }

// Uint32n returns pseudorandom uint32 in the range [0...maxN).
// It is safe calling this function from concurrent goroutines.
func Uint32n(maxN uint32) uint32 { return fastrand.Uint32n(maxN) }

// Probability 小于prob的概率, prob is in the range [0,1.0)
func Probability(prob float64) bool {
	if prob > float64(Uint32()) {
		return true
	}
	return false
}

// Perm
// returns a random permutation of the range [0,n).
func Perm(n uint32) []uint32 {
	m := make([]uint32, n)
	for i := uint32(1); i < n; i++ {
		j := Uint32n(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}

// Shuffle pseudo-randomizes the order of elements.
// 	n: is the number of elements. Shuffle panics if n < 0.
// 	swap: swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(Uint32n(uint32(i + 1)))
		swap(i, j)
	}

	for ; i > 0; i-- {
		j := int(Uint32n(uint32(i + 1)))
		swap(i, j)
	}
}

// Random
// return random string from string slice
func Random(ss []string) []string {
	for i := uint32(len(ss)) - 1; i > 0; i-- {
		num := Uint32n(i + 1)
		ss[i], ss[num] = ss[num], ss[i]
	}
	return ss[:len(ss):len(ss)]
}
