package cache

import (
	"runtime"
	"time"

	"github.com/pkg/errors"
)

// 定时清理过期数据
func initJanitor(p *cache) error {
	interval := p.opts.ClearTime
	if interval < defaultMinExpiration {
		return errors.Wrapf(ErrClearTime, "过期时间(%s)小于最小过期时间(%s)", interval, defaultMinExpiration)
	}

	if p.janitor == nil {
		runtime.SetFinalizer(p, stopJanitor)
	} else {
		stopJanitor(p)
	}

	runJanitor(p, interval)
	return nil
}

func stopJanitor(c *cache) {
	c.janitor.stop <- true
}

func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *cache) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}
