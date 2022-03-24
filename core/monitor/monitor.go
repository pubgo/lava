package monitor

import (
	"context"
	"time"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/xerror"
)

var logs = logging.Component("monitor")

func Register(run func(num int, log *logging.Logger) bool) {

}

func Monitor(timeout time.Duration, run func(), errFn func(err error)) {
	if timeout <= 0 {
		panic("[Monitor] [timeout] should not be less than zero")
	}

	checkFn(run, "[Monitor] [run] is nil")
	checkFn(errFn, "[Monitor] [errFn] is nil")

	var done = make(chan struct{})
	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			logutil.ErrTry(logs.L(), func() { errFn(err) }, logutil.FuncStack(run))
		})

		run()
		close(done)
	}()

	for {
		select {
		case <-time.After(timeout):
			logutil.ErrTry(logs.L(), func() { errFn(context.DeadlineExceeded) }, logutil.FuncStack(run))
		case <-done:
			return
		}
	}
}

func logErr(fn interface{}, err xerror.XErr) {
	logs.WithErr(err).With(logutil.FuncStack(fn)).Error(err.Error())
}

func checkFn(fn interface{}, msg string) {
	if fn == nil {
		panic(msg)
	}
}
