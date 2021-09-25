package cache

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/pubgo/lug/logger"

	"github.com/pkg/errors"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"
)

type cacheImpl struct {
	sync.RWMutex
	store      IStore
	onDataLoad func(string) (interface{}, error)
	onEvicted  func(k string, v []byte)
	bufSize    atomic.Uint32
	count      atomic.Uint32
	cfg        Cfg
	sf         singleflight.Group
	janitor    *janitor
}

func (p *cacheImpl) deleteExpired() {
	_, _, _ = p.sf.Do("delete_expired", func() (interface{}, error) {
		_ = p.store.DeleteExpired()
		return nil, nil
	})
}

// 过期时间处理, 防止缓存雪崩
func (p *cacheImpl) expiredHandle(d time.Duration) time.Duration {
	return d + time.Duration(rand.Intn(int(defaultMinExpiration)))
}

// 缓存key长度处理
func (p *cacheImpl) checkKey(k string) error {
	if l := uint32(len(k)); l > p.cfg.MaxKeySize || l < defaultMinKeySize {
		return xerror.Fmt("key:%s", k)
	}
	return nil
}

// 缓存过期时间处理
func (p *cacheImpl) checkExpiration(d time.Duration) error {
	if d > p.cfg.MaxExpiration || d < defaultMinExpiration {
		return xerror.Fmt("expired time error, dur:%s", d)
	}
	return nil
}

func (p *cacheImpl) set(k string, x interface{}, dur time.Duration) (err error) {
	var val1 []byte
	if x != nil {
		val1, err = json.Marshal(x)
		if err != nil {
			return
		}
	}

	p.bufSize.Add(uint32(len(k) + len(val1)))

	// 检测缓存是否过期
	// 检测缓存是否存在
	val, _, err := p.store.GetExpired(k)
	if err != nil {
		p.bufSize.Sub(uint32(len(k) + len(val)))
	}

	return xerror.Wrap(p.store.Set(k, val1, dur))
}

func (p *cacheImpl) Set(k string, x interface{}, d time.Duration) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(p.checkKey(k))
	xerror.Panic(p.checkExpiration(d))

	if d == defaultExpiration {
		d = p.cfg.DefaultExpiration
	}

	// 缓存value为nil, 设置最小缓存时间, 防止缓存穿透
	if x == nil {
		d = defaultMinExpiration
	}

	// 缓存大小检测
	if p.bufSize.Load() > p.cfg.MaxBufSize {
		return errors.Wrapf(ErrBufExceeded, "bufSize:%d, MaxBufSize:%d", p.bufSize.Load(), p.cfg.MaxBufSize)
	}

	var dur = p.expiredHandle(d)

	p.Lock()
	defer p.Unlock()
	return p.set(k, x, dur)
}

func (p *cacheImpl) SetDefault(k string, x interface{}) error {
	return p.Set(k, x, p.cfg.DefaultExpiration)
}

func (p *cacheImpl) getSet(key string, d time.Duration, fn ...getCallback) (_ []byte, gErr error) {
	defer xerror.RespErr(&gErr)

	if err := p.checkKey(key); err != nil {
		return nil, err
	}

	dataLoad := p.onDataLoad
	if len(fn) > 0 {
		dataLoad = fn[0]
	}

	dt, expireAt, err := p.store.GetExpired(key)
	// 缓存不存在或者过期, 就从数据库重新获取一份
	if err == ErrNotFound || time.Since(expireAt) > 0 {
		if dataLoad == nil {
			if err == ErrNotFound {
				return nil, ErrNotFound
			} else {
				return nil, xerror.Wrap(ErrKeyExpired, "key:%s", key)
			}
		}

		var val interface{}
		xerror.Panic(fx.Timeout(dataLoadTimeOutTime, func() {
			val, err = dataLoad(key)
			xerror.Panic(err)
		}))

		return dt, xerror.Wrap(p.Set(key, dt, d))
	}

	return dt, xerror.Wrap(err)
}

func (p *cacheImpl) GetSet(key string, d time.Duration, fn ...getCallback) ([]byte, error) {
	return p.getSet(key, d, fn...)
}

func (p *cacheImpl) Get(key string, fn ...getCallback) ([]byte, error) {
	return p.getSet(key, p.cfg.DefaultExpiration, fn...)
}

func (p *cacheImpl) Delete(k string) error {
	if err := p.checkKey(k); err != nil {
		return err
	}

	p.Lock()
	defer p.Unlock()
	return p.store.Delete(k)
}

func (p *cacheImpl) BufSize() uint32 {
	return p.bufSize.Load()
}

func (p *cacheImpl) OnDataLoad(f func(string) (interface{}, error)) {
	p.Lock()
	defer p.Unlock()

	p.onDataLoad = f
}

func (p *cacheImpl) defaultEvicted(k string, v []byte) {
	p.bufSize.Sub(uint32(len(k) + len(v)))

	onEvicted := p.onEvicted
	if onEvicted == nil {
		return
	}

	logger.Logs(func() { onEvicted(k, v) })
}

func (p *cacheImpl) OnEvicted(f func(k string, v []byte)) {
	p.Lock()
	defer p.Unlock()

	p.onEvicted = f
}

// Close 关闭清理
func (p *cacheImpl) Close() error {
	p.Lock()
	defer p.Unlock()

	stopJanitor(p)

	p.onDataLoad = nil
	p.onEvicted = nil
	p.bufSize.Store(0)
	p.cfg = Cfg{}
	p.sf = singleflight.Group{}
	return p.store.Close()
}
