package rsync

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pubgo/xlog"
	"go.uber.org/atomic"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Range(min, max int) int {
	return min + rand.Intn(max-min)
}

func NewWaiter() *Waiter {
	return &Waiter{
		changed: make(map[string]*atomic.Bool),
		skip:    make(map[string]*atomic.Uint32),
	}
}

type Waiter struct {
	mu      sync.Mutex
	changed map[string]*atomic.Bool
	skip    map[string]*atomic.Uint32
}

func (t *Waiter) check(key string) {
	if _, ok := t.changed[key]; !ok {
		t.skip[key] = atomic.NewUint32(0)
		t.changed[key] = atomic.NewBool(false)
	}
}

func (t *Waiter) Report(key string, c *atomic.Bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.check(key)

	if c.Load() {
		t.changed[key].Store(true)
		t.skip[key].Store(0)
		return
	}

	t.changed[key].Store(false)
	t.skip[key].Inc()
}

func (t *Waiter) Skip(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.check(key)

	if t.changed[key].Load() || t.skip[key].Load() == 0 {
		return false
	}

	if t.skip[key].Load() > uint32(Range(5, 120)) {
		t.skip[key].Store(0)
		xlog.Debugf("no skip: %s", key)
		return false
	}

	t.skip[key].Inc()
	xlog.Debugf("skip: %s", key)
	return true
}
