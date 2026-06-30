package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

// TestNotifyShutdownCancelsOnSignal 验证首个信号取消 context，第二个信号触发强制退出回调。
// 使用 SIGUSR1（不会终止测试进程）作为受控信号。
func TestNotifyShutdownCancelsOnSignal(t *testing.T) {
	forced := make(chan struct{})
	ctx := notifyShutdown(context.Background(), []os.Signal{syscall.SIGUSR1}, func() {
		close(forced)
	})

	// 第一个信号：context 应被取消
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send first signal: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatalf("context not cancelled after first signal")
	}

	// 第二个信号：应触发强制退出回调
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send second signal: %v", err)
	}

	select {
	case <-forced:
	case <-time.After(2 * time.Second):
		t.Fatalf("force-exit callback not invoked after second signal")
	}
}

// TestShutdownSignalsExcludeUncatchable 锁定回归：不可捕获的信号不应出现在列表里。
func TestShutdownSignalsExcludeUncatchable(t *testing.T) {
	for _, sig := range shutdownSignals {
		if sig == os.Kill {
			t.Fatalf("os.Kill (SIGKILL) cannot be caught and must not be registered")
		}
		if sig == syscall.SIGSTOP {
			t.Fatalf("SIGSTOP cannot be caught and must not be registered")
		}
	}
}
