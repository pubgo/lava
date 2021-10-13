package lavax

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// #nosec G103
// UnsafeString returns a string pointer without allocation
func UnsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// #nosec G103
// UnsafeBytes returns a byte pointer without allocation
func UnsafeBytes(s string) (bs []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return
}

// CopyString copies a string to make it immutable
func CopyString(s string) string {
	return string(UnsafeBytes(s))
}

// CopyBytes copies a slice to make it immutable
func CopyBytes(b []byte) []byte {
	tmp := make([]byte, len(b))
	copy(tmp, b)
	return tmp
}

const (
	uByte = 1 << (10 * iota)
	uKilobyte
	uMegabyte
	uGigabyte
	uTerabyte
	uPetabyte
	uExabyte
)

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)
	switch {
	case bytes >= uExabyte:
		unit = "EB"
		value = value / uExabyte
	case bytes >= uPetabyte:
		unit = "PB"
		value = value / uPetabyte
	case bytes >= uTerabyte:
		unit = "TB"
		value = value / uTerabyte
	case bytes >= uGigabyte:
		unit = "GB"
		value = value / uGigabyte
	case bytes >= uMegabyte:
		unit = "MB"
		value = value / uMegabyte
	case bytes >= uKilobyte:
		unit = "KB"
		value = value / uKilobyte
	case bytes >= uByte:
		unit = "B"
	default:
		return "0B"
	}
	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}

// Copy creates an identical copy of x
func Copy(x []byte) []byte { return x[:len(x):len(x)] }

// InsideTest returns true inside a Go test
func InsideTest() bool {
	return len(os.Args) > 1 && strings.HasSuffix(os.Args[0], ".test") &&
		strings.HasPrefix(os.Args[1], "-test.")
}

// String internals from reflect
type String struct {
	Data unsafe.Pointer
	Len  int
}

// Slice internals from reflect
type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// BtoU4 converts byte slice to integer slice
func BtoU4(b []byte) (i []uint32) {
	I := (*Slice)(unsafe.Pointer(&i))
	B := (*Slice)(unsafe.Pointer(&b))
	I.Data = B.Data
	I.Len = B.Len >> 2
	I.Cap = I.Len
	return
}

// U4toB converts integer slice to byte slice
func U4toB(i []uint32) (b []byte) {
	I := (*Slice)(unsafe.Pointer(&i))
	B := (*Slice)(unsafe.Pointer(&b))
	B.Data = I.Data
	B.Len = I.Len << 2
	B.Cap = B.Len
	return
}

// U4toU8 converts uint32 slice to uint64 slice
func U4toU8(i []uint32) (k []uint64) {
	I := (*Slice)(unsafe.Pointer(&i))
	K := (*Slice)(unsafe.Pointer(&k))
	K.Data = I.Data
	K.Len = I.Len >> 1
	K.Cap = K.Len
	return
}

// U8toU4 converts uint64 slice to uint32 slice
func U8toU4(i []uint64) (k []uint32) {
	I := (*Slice)(unsafe.Pointer(&i))
	K := (*Slice)(unsafe.Pointer(&k))
	K.Data = I.Data
	K.Len = I.Len << 1
	K.Cap = K.Len
	return
}

// BtoU8 converts byte slice to integer slice
func BtoU8(b []byte) (i []uint64) {
	I := (*Slice)(unsafe.Pointer(&i))
	B := (*Slice)(unsafe.Pointer(&b))
	I.Data = B.Data
	I.Len = B.Len >> 3
	I.Cap = I.Len
	return
}

// U8toB converts integer slice to byte slice
func U8toB(i []uint64) (b []byte) {
	I := (*Slice)(unsafe.Pointer(&i))
	B := (*Slice)(unsafe.Pointer(&b))
	B.Data = I.Data
	B.Len = I.Len << 3
	B.Cap = B.Len
	return
}

// BtoS converts byte slice to string
func BtoS(b []byte) (s string) {
	B := (*Slice)(unsafe.Pointer(&b))
	S := (*String)(unsafe.Pointer(&s))
	S.Data = B.Data
	S.Len = B.Len
	return
}

// StoB converts string to byte slice
func StoB(s string) (b []byte) {
	B := (*Slice)(unsafe.Pointer(&b))
	S := (*String)(unsafe.Pointer(&s))
	B.Data = S.Data
	B.Len = S.Len
	B.Cap = B.Len
	return
}

// StoU4 converts string to integer slice
func StoU4(s string) (i []uint32) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*String)(unsafe.Pointer(&s))
	I.Data = S.Data
	I.Len = S.Len >> 2
	I.Cap = I.Len
	return
}

// U4toS converts integer slice to string
func U4toS(i []uint32) (s string) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*String)(unsafe.Pointer(&s))
	S.Data = I.Data
	S.Len = I.Len << 2
	return
}

// StoU8 converts string to integer slice
func StoU8(s string) (i []uint64) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*String)(unsafe.Pointer(&s))
	I.Data = S.Data
	I.Len = S.Len >> 3
	I.Cap = I.Len
	return
}

// U8toS converts integer slice to string
func U8toS(i []uint64) (s string) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*String)(unsafe.Pointer(&s))
	S.Data = I.Data
	S.Len = I.Len << 3
	return
}

const (
	// StrSize is size of a string variable
	StrSize = int(unsafe.Sizeof(""))
	// SliceSize is size of a slice variable
	SliceSize = int(unsafe.Sizeof([]byte{}))
)

// B2Strs converts byte slice to String slice
func B2Strs(b []byte) (ss []String) {
	B := (*Slice)(unsafe.Pointer(&b))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = B.Data
	S.Len = B.Len / StrSize
	S.Cap = S.Len
	return
}

// U4toStrs converts integer slice to String slice
func U4toStrs(i []uint32) (ss []String) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = I.Data
	S.Len = 4 * I.Len / StrSize
	S.Cap = S.Len
	return
}

// U8toStrs converts integer slice to String slice
func U8toStrs(i []uint64) (ss []String) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = I.Data
	S.Len = 8 * I.Len / StrSize
	S.Cap = S.Len
	return
}

// BtoSlices converts byte slice to Slice list
func BtoSlices(b []byte) (ss []Slice) {
	B := (*Slice)(unsafe.Pointer(&b))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = B.Data
	S.Len = B.Len / SliceSize
	S.Cap = S.Len
	return
}

// U4toSlices converts integer slice to Slice list
func U4toSlices(i []uint32) (ss []Slice) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = I.Data
	S.Len = 4 * I.Len / SliceSize
	S.Cap = S.Len
	return
}

// U8toSlices converts integer slice to Slice list
func U8toSlices(i []uint64) (ss []Slice) {
	I := (*Slice)(unsafe.Pointer(&i))
	S := (*Slice)(unsafe.Pointer(&ss))
	S.Data = I.Data
	S.Len = 8 * I.Len / SliceSize
	S.Cap = S.Len
	return
}

// CmpS returns -1 for a < b, 0 for a = b, and 1 for a > b lexicographically
func CmpS(a, b string) (r int) {
	n, k := len(a), len(b)
	if n > k {
		n = k
		r++
	} else if n < k {
		r--
	}

	for i := 0; i < n; i++ {
		x, y := a[i], b[i]
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return
}

// CmpB returns -1 for a < b, 0 for a = b, and 1 for a > b lexicographically
func CmpB(a, b []byte) int {
	return CmpS(BtoS(a), BtoS(b))
}
