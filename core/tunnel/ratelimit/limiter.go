// Package ratelimit 提供令牌桶速率限制，供 gateway 代理与 P2P 信令复用。
package ratelimit

import (
	"sync"
	"time"
)

// Limiter 按 key 维度的令牌桶限流器（key 未配置时使用 defaultLimit）。
type Limiter struct {
	limits       map[string]int
	buckets      map[string]*TokenBucket
	mu           sync.RWMutex
	defaultLimit int
}

// New 创建限流器；defaultLimit 为每个 key 的默认每秒配额。
func New(defaultLimit int) *Limiter {
	if defaultLimit <= 0 {
		defaultLimit = 1
	}
	return &Limiter{
		limits:       make(map[string]int),
		buckets:      make(map[string]*TokenBucket),
		defaultLimit: defaultLimit,
	}
}

// SetLimit 设置指定 key 的每秒配额。
func (rl *Limiter) SetLimit(key string, limit int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limits[key] = limit
	rl.buckets[key] = NewTokenBucket(limit, limit)
}

// GetLimit 返回 key 的配额（未配置时返回默认值）。
func (rl *Limiter) GetLimit(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	if limit, ok := rl.limits[key]; ok {
		return limit
	}
	return rl.defaultLimit
}

// Allow 检查 key 是否允许通过一次请求。
func (rl *Limiter) Allow(key string) bool {
	rl.mu.RLock()
	bucket, ok := rl.buckets[key]
	rl.mu.RUnlock()
	if !ok {
		rl.mu.Lock()
		bucket, ok = rl.buckets[key]
		if !ok {
			limit := rl.defaultLimit
			if l, ok := rl.limits[key]; ok {
				limit = l
			}
			bucket = NewTokenBucket(limit, limit)
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}
	return bucket.Allow()
}

// TokenBucket 令牌桶。
type TokenBucket struct {
	capacity   int
	rate       int
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket 创建令牌桶；capacity 与 rate 通常为同一值（允许突发一整秒配额）。
func NewTokenBucket(capacity, rate int) *TokenBucket {
	if capacity <= 0 {
		capacity = 1
	}
	if rate <= 0 {
		rate = capacity
	}
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

// Allow 消耗一个令牌；成功返回 true。
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tokensToAdd := now.Sub(tb.lastRefill).Seconds() * float64(tb.rate)
	tb.tokens += tokensToAdd
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}
