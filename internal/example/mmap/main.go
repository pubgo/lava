package main

import (
	mm "github.com/edsrzf/mmap-go"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	f := openFile(os.O_RDWR | os.O_CREATE | os.O_TRUNC)
	defer f.Close()

	mmap, err := mm.Map(f, mm.RDONLY, 0)
	if err != nil {
		zap.S().Errorf("error mapping: %s", err)
	}

	if err := mmap.Unmap(); err != nil {
		zap.S().Errorf("error unmapping: %s", err)
	}
}

// Open will mmap a file to a byte slice of data.
func Open(path string, writable bool) (data []byte, err error) {
	flag, prot := os.O_RDONLY, syscall.PROT_READ
	if writable {
		flag, prot = os.O_RDWR, syscall.PROT_READ|syscall.PROT_WRITE
	}
	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() > 0 {
		return syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), prot, syscall.MAP_SHARED)
	}
	return nil, nil
}

// Close releases the data. Don't read the data after running this operation
// otherwise your f*cked.
func Close(data []byte) error {
	if len(data) > 0 {
		return syscall.Munmap(data)
	}
	return nil
}

var testPath = filepath.Join(os.TempDir(), "testdata")

func openFile(flags int) *os.File {
	f, err := os.OpenFile(testPath, flags, 0666)
	if err != nil {
		panic(err.Error())
	}
	return f
}
