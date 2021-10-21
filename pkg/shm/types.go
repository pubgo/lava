package shm

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

var (
	errNotEnough = errors.New("span capacity is not enough")
)

func path(name string) string {
	return "MosnConfigPath" + string(os.PathSeparator) + fmt.Sprintf("mosn_shm_%s", name)
}

// check if given path match the required size
// return error if path exists and size not match
func checkConsistency(path string, size int) error {
	if info, err := os.Stat(path); err == nil {
		if info.Size() != int64(size) {
			return errors.New(fmt.Sprintf("mmap target path %s exists and its size %d mismatch %d", path, info.Size(), size))
		}
	}
	return nil
}

type Span struct {
	origin []byte
	name   string

	data   uintptr
	offset int
	size   int
}

func NewShmSpan(name string, data []byte) *Span {
	return &Span{
		name:   name,
		origin: data,
		data:   uintptr(unsafe.Pointer(&data[0])),
		size:   len(data),
	}
}

func (s *Span) Alloc(size int) (uintptr, error) {
	if s.offset+size > s.size {
		return 0, errNotEnough
	}

	ptr := s.data + uintptr(s.offset)
	s.offset += size
	return ptr, nil
}

func (s *Span) Data() uintptr {
	return s.data
}

func (s *Span) Origin() []byte {
	return s.origin
}
