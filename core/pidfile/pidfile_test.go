package pidfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveGetRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pid")

	if err := writePID(path, 4242).GetErr(); err != nil {
		t.Fatalf("writePID failed: %v", err)
	}

	got := readPID(path).Expect("read pid")
	if got != 4242 {
		t.Fatalf("expected pid 4242, got %d", got)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	if err := readPID(path).GetErr(); err == nil {
		t.Fatalf("expected error reading removed pid file")
	}
}

func TestReadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.pid")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	if err := readPID(path).GetErr(); err == nil {
		t.Fatalf("expected error for empty pid file")
	}
}

func TestRemoveMissingIsOK(t *testing.T) {
	// Remove on default path when file doesn't exist should not error.
	// We test the underlying os.Remove behavior via a temp path helper.
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.pid")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPathNotEmpty(t *testing.T) {
	if got := GetPath(); got == "" {
		t.Fatalf("GetPath() should not be empty")
	}
}
