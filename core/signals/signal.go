// Package signals 将操作系统的关停信号转换为 context 取消，用于服务优雅退出。
//
// 典型用法是在进程入口获取一个随关停信号取消的根 context：
//
//	ctx := signals.Context()
//	app.Run(ctx) // 收到 SIGTERM/SIGINT 等信号时 ctx 被取消
//
// 行为约定：
//   - 首次收到关停信号：取消返回的 context，触发优雅退出流程；
//   - 再次收到关停信号：强制 os.Exit(1)，避免优雅退出卡死时无法终止；
//   - Context 在整个进程生命周期内只能调用一次，重复调用会 panic。
package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/v2/log"
)

const Name = "signals"

var logger = log.GetLogger(Name)

// Inspired by
// https://github.com/kubernetes-sigs/controller-runtime/blob/8499b67e316a03b260c73f92d0380de8cd2e97a1/pkg/manager/signals/signal.go#L25
var onlyOneSignalHandler = make(chan struct{})

// shutdownSignals 是触发优雅关停的信号集合。
// 注意：SIGKILL(os.Kill) 与 SIGSTOP 无法被进程捕获，因此不在此列表中。
var shutdownSignals = []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt}

// Context 返回一个会在收到关停信号时被取消的 context。
//
// 该函数在整个进程内只能调用一次，重复调用会 panic（防止注册多个信号处理器）。
// 收到第二次关停信号时进程会强制退出。
func Context() context.Context {
	close(onlyOneSignalHandler) // 重复调用会 panic

	return notifyShutdown(context.Background(), shutdownSignals, func() { os.Exit(1) })
}

// notifyShutdown 注册信号监听并返回随首个信号取消的 context；
// 收到第二个信号时调用 onForceExit。onForceExit 抽出为参数以便测试。
func notifyShutdown(parent context.Context, sigs []os.Signal, onForceExit func()) context.Context {
	ctx, cancel := context.WithCancel(parent)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, sigs...)

	go func() {
		sig := <-ch
		logger.Info().Msgf("cancelling context, received signal:%s", sig)
		cancel()

		sig = <-ch
		logger.Info().Msgf("os exit, received twice signal:%s", sig)
		onForceExit()
	}()

	return ctx
}
