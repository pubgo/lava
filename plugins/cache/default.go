package cache

import "time"

var defaultCache *cacheImpl

func init() {
	defaultCache = nil
}

func SetDefault(c *cacheImpl) {
	defaultCache = c
}

// Get 全局获取缓存
func Get(k string) (interface{}, error) { return defaultCache.Get(k) }

// Set 全局存储缓存
func Set(k string, x interface{}, ds time.Duration) error { return defaultCache.Set(k, x, ds) }

// GetSet 自定义过期时间获取或者设置缓存
func GetSet(k string, d time.Duration, fn ...getCallback) (interface{}, error) {
	return defaultCache.GetSet(k, d, fn...)
}

// Delete 全局删除缓存
func Delete(k string) error { return defaultCache.Delete(k) }
