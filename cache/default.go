package cache

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"
)

const (
	// DefaultExpiration 默认过期时间
	defaultExpiration time.Duration = 0
	// 最小过期时间
	defaultMinExpiration = time.Second * 2
	// 最大过期时间10m
	defaultMaxExpiration = time.Minute * 10

	// key长度限制
	// key最小长度
	defaultMinKeySize = 5
	// key最大长度
	defaultMaxKeySize = 65535

	// DataLoadTimeOutTime 默认的数据加载过期时间
	dataLoadTimeOutTime = time.Second * 15

	// 默认过期定时清理时间2m
	defaultClearTime = time.Minute * 2
)

type cache struct {
	sync.RWMutex
	onDataLoad func([]byte) ([]byte, error)
	onEvicted  func(k, v []byte)
	bufSize    atomic.Uint64
	opts       options
	sf         singleflight.Group
	janitor    *janitor
}

func (p *cache) Options() options {
	return p.opts
}

func (p *cache) Init(opts ...Option) error {
	p.Lock()
	defer p.Unlock()

	opt := p.opts

	for _, o := range opts {
		o(&opt)
	}

	// 最大缓存大小判断
	if opt.MaxBufSize > DefaultMaxBufSize || opt.MaxBufSize < DefaultMinBufSize {
		return errors.Wrapf(ErrBufSize, "ErrBufSize:%d, min:%d, max:%d", opt.MaxBufSize, DefaultMinBufSize, DefaultMaxBufSize)
	}

	// 最大过期时间判断
	if opt.MaxExpiration > defaultMaxExpiration || opt.MaxExpiration < defaultMinExpiration {
		return errors.Wrapf(ErrExpiration, "MaxExpiration:%d, min:%d, max:%d", opt.DefaultExpiration, defaultMinExpiration, defaultMaxExpiration)
	}

	// 默认过期时间判断
	if opt.DefaultExpiration < defaultMinExpiration || opt.DefaultExpiration > opt.MaxExpiration {
		return errors.Wrapf(ErrExpiration, "DefaultExpiration:%d, min:%d, max:%d", opt.DefaultExpiration, defaultMinExpiration, defaultMaxExpiration)
	}

	// key 长度判断
	if opt.MaxKeySize > defaultMaxKeySize || opt.MaxKeySize < defaultMinKeySize {
		return errors.Wrapf(ErrKeyLength, "MaxKeySize:%d, min:%d, max:%d", opt.MaxKeySize, defaultMinKeySize, defaultMaxKeySize)
	}

	// 存储器判断
	if opt.store == nil {
		return errors.Wrapf(ErrStore, "store is nil")
	}

	if err := initJanitor(p); err != nil {
		return err
	}

	// 过期驱逐设置
	opt.store.OnEvicted(p.onDftEvicted)

	p.opts = opt

	return nil
}

func (p *cache) init() {
	p.opts.ClearTime = defaultClearTime
	p.opts.DataLoadTime = dataLoadTimeOutTime
	p.opts.MaxBufSize = DefaultMaxBufSize
	p.opts.MaxExpiration = defaultMaxExpiration
	p.opts.MaxKeySize = defaultMaxKeySize
	p.opts.DefaultExpiration = -1
	p.opts.store = &noopStore{}
}

func (p *cache) deleteExpired() {
	_, _, _ = p.sf.Do("delete_expired", func() (interface{}, error) {
		_ = p.opts.store.DeleteExpired()
		return nil, nil
	})
}

// 过期时间处理
func (p *cache) expiredHandle(d time.Duration) time.Duration {
	return d + time.Duration(rand.Intn(int(defaultMinExpiration)))
}

func (p *cache) checkKey(k string) error {
	if l := uint64(len(k)); l > p.opts.MaxKeySize || l < defaultMinKeySize {
		return errors.Wrapf(ErrKeyLength, "key:%s", k)
	}
	return nil
}

func (p *cache) checkExpiration(d time.Duration) error {
	if d > p.opts.MaxExpiration || d < defaultMinExpiration {
		return errors.Wrapf(ErrExpiration, "expired time error, dur:%s", d)
	}
	return nil
}

func (p *cache) set(k []byte, x []byte, d time.Duration) error {
	if d == defaultExpiration {
		d = p.opts.DefaultExpiration
	}

	if x == nil {
		d = defaultMinExpiration
	}

	size := uint64(len(k) + len(x))

	// 全局缓存检查
	gBufSize := GlobalBufSize.Add(size)
	if gBufSize > uint64(GlobalMaxBufExpand) {
		GlobalBufSize.Sub(size)
		return errors.Wrapf(ErrBufExceeded, "GlobalBufSize:%d, GlobalMaxBufExpand:%f", gBufSize, GlobalMaxBufExpand)
	}

	if gBufSize > GlobalMaxBufSize {
		// 清理过期
		go p.deleteExpired()
	}

	bufSize := p.bufSize.Add(size)
	// 超过了本实例的缓存
	if bufSize > p.opts.MaxBufSize {
		p.bufSize.Sub(size)
		return errors.Wrapf(ErrBufExceeded, "bufSize:%d, MaxBufSize:%d", bufSize, p.opts.MaxBufSize)
	}

	size1 := uint64(0)
	val, _, err := p.opts.store.GetWithExpiration(k)
	if err == nil {
		size1 = uint64(len(k) + len(val))
	}

	if err := p.opts.store.Set(k, x, p.expiredHandle(d)); err != nil {
		return err
	}

	GlobalBufSize.Sub(size1)
	p.bufSize.Sub(size1)

	return nil
}

func (p *cache) Set(k string, x interface{}, d time.Duration) error {
	if err := p.checkKey(k); err != nil {
		return err
	}

	if err := p.checkExpiration(d); err != nil {
		return err
	}

	p.Lock()
	defer p.Unlock()
	return p.set(k, x, d)
}

func (p *cache) SetDefault(k string, x interface{}) error {
	return p.Set(k, x, p.opts.DefaultExpiration)
}

func (p *cache) getSet(key string, d time.Duration, fn ...getCallback) ([]byte, error) {
	if err := p.checkKey(key); err != nil {
		return nil, err
	}

	dataLoad := p.onDataLoad
	if len(fn) > 0 {
		dataLoad = fn[0]
	}

	dt, err := p.opts.store.Get(key)
	if err != ErrKeyNotFound {
		return dt, err
	}

	if dataLoad == nil {
		return nil, ErrKeyNotFound
	}

	// key不存在
	var ch = make(chan error, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				ch <- fmt.Errorf("%s", err)
			}
		}()

		dt, err = dataLoad(key)
		ch <- err
	}()

	select {
	case <-time.After(dataLoadTimeOutTime):
		return nil, ErrDataLoadTimeOut
	case err = <-ch:
		if err != nil {
			return nil, err
		}
	}

	var expired = defaultMinExpiration
	if dt != nil {
		expired = d
	}

	return dt, p.Set(key, dt, expired)
}

func (p *cache) GetSet(key string, d time.Duration, fn ...getCallback) ([]byte, error) {
	return p.getSet(key, d, fn...)
}

func (p *cache) Get(key string, fn ...func([]byte) ([]byte, error)) ([]byte, error) {
	return p.getSet(key, p.opts.DefaultExpiration, fn...)
}

func (p *cache) Delete(k string) error {
	if err := p.checkKey(k); err != nil {
		return err
	}

	p.Lock()
	defer p.Unlock()
	return p.opts.store.Delete(k)
}

func (p *cache) BufSize() uint64 {
	return p.bufSize.Load()
}

func (p *cache) OnDataLoad(f func([]byte) ([]byte, error)) {
	p.Lock()
	defer p.Unlock()

	p.onDataLoad = f
}

func (p *cache) onDftEvicted(k, v []byte) {
	size := uint64(len(k) + len(v))
	GlobalBufSize.Sub(size)
	p.bufSize.Sub(size)

	if p.onEvicted != nil {
		onEvicted := p.onEvicted
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println(err)
				}
			}()
			onEvicted(k, v)
		}()
	}
}

func (p *cache) OnEvicted(f func(k, v []byte)) {
	p.Lock()
	defer p.Unlock()

	p.onEvicted = f
}

// Close 关闭清理
func (p *cache) Close() error {
	p.Lock()
	defer p.Unlock()

	GlobalBufSize.Sub(p.bufSize.Load())
	stopJanitor(p)

	p.onDataLoad = nil
	p.onEvicted = nil
	p.bufSize.Store(0)
	p.opts = options{}
	p.sf = singleflight.Group{}
	return p.opts.store.Close()
}

func newPCache(opts ...Option) (*cache, error) {
	var c = &cache{}
	c.init()
	return c, c.Init(opts...)
}

var defaultPCache *cache

func SetDefaultCache(c *cache) {
	defaultPCache = c
}

// Get 全局获取缓存
func Get(k string) (interface{}, error) { return defaultPCache.Get(k) }

// Set 全局存储缓存
func Set(k string, x interface{}, ds ...time.Duration) error { return defaultPCache.Set(k, x, ds...) }

// GetSet 自定义过期时间获取或者设置缓存
func GetSet(k string, d time.Duration, fn ...getCallback) (interface{}, error) {
	return defaultPCache.GetSet(k, d, fn...)
}

// Delete 全局删除缓存
func Delete(k string) error { return defaultPCache.Delete(k) }
