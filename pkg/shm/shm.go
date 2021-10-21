package shm

import (
	"os"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
)

func Alloc(name string, size int) (*Span, error) {
	path := path(name)

	os.MkdirAll(filepath.Dir(path), 0755)

	// check consistency
	if err := checkConsistency(path, size); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	if err := f.Truncate(int64(size)); err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {
		return nil, err
	}

	// lock mmap data to avoid I/O page fault
	err = syscall.Mlock(data)
	if err != nil {
		zap.S().Warnf("failed to mlock memory from mmap, please check the RLIMIT_MEMLOCK:%s\n", err)
	}

	return NewShmSpan(name, data), nil
}

func Free(span *Span) error {
	Clear(span.name)
	return syscall.Munmap(span.origin)
}

func Clear(name string) error {
	return os.Remove(path(name))
}
