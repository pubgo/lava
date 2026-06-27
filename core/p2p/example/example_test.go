package main_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func exampleDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func TestExampleLocal(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-mode=local")
	cmd.Dir = exampleDir(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("local example: %v\n%s", err, out)
	}
	t.Logf("output:\n%s", out)
}

func TestExampleAll(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".", "-mode=all", "-gateway-addr=127.0.0.1:27001")
	cmd.Dir = exampleDir(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("all example: %v\n%s", err, out)
	}
	t.Logf("output:\n%s", out)
}
