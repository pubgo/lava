package syncx

import (
	"sync"
	"sync/atomic"
	_ "unsafe"

	"github.com/pubgo/lava/internal/pkg/fastrand"
)

const DefaultConcurrent = 16

//go:linkname state sync.(*WaitGroup).state
func state(*sync.WaitGroup) (*uint64, *uint32)

type WaitGroup struct {
	wg         sync.WaitGroup
	Concurrent uint32
}

func (t *WaitGroup) Count() uint32 {
	count, _ := state(&t.wg)
	return uint32(atomic.LoadUint64(count) >> 32)
}

func (t *WaitGroup) check() {
	if t.Concurrent == 0 {
		t.Concurrent = DefaultConcurrent
	}

	// 阻塞, 等待任务处理完毕
	// 采样率(10), 打印log
	if t.Count() >= t.Concurrent && fastrand.Sampling(10) {
		logs.S().Warnw("WaitGroup current concurrent number exceeds the maximum concurrent number of the system",
			"current", t.Count(), "maximum", t.Concurrent)
	}
}

func (t *WaitGroup) Inc()          { t.check(); t.wg.Add(1) }
func (t *WaitGroup) Dec()          { t.wg.Done() }
func (t *WaitGroup) Done()         { t.wg.Done() }
func (t *WaitGroup) Wait()         { t.wg.Wait() }
func (t *WaitGroup) Add(delta int) { t.check(); t.wg.Add(delta) }
