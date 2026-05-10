package ttlcache

import (
	"sync"
	"time"
)

// cacheItem 缓存项
type cacheItem struct {
	createdAt time.Time
	value     any
}

// TimedCache 带过期时间的缓存
type TimedCache struct {
	mutex    sync.RWMutex
	interval time.Duration
	stopChan chan struct{}
	items    map[string]cacheItem
}

// NewTimedCache 创建带过期时间的缓存
func NewTimedCache(interval time.Duration) *TimedCache {
	cache := &TimedCache{
		interval: interval,
		stopChan: make(chan struct{}),
		items:    make(map[string]cacheItem),
	}

	go func() {
		ticker := time.NewTicker(cache.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cache.DeleteExpired()
			case <-cache.stopChan:
				return
			}
		}
	}()

	return cache
}

// Set 设置缓存值
func (c *TimedCache) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[key] = cacheItem{
		createdAt: time.Now(),
		value:     value,
	}
}

// Get 获取缓存值，返回值和是否存在
func (c *TimedCache) Get(key string) (any, bool) {
	c.mutex.RLock()
	item, ok := c.items[key]
	c.mutex.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Since(item.createdAt) >= c.interval {
		// 过期时删除
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()
		return nil, false
	}

	return item.value, true
}

// DeleteExpired 删除所有过期的缓存项
func (c *TimedCache) DeleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, item := range c.items {
		if time.Since(item.createdAt) > c.interval {
			delete(c.items, key)
		}
	}
}

// StopCleanup 停止清理协程
func (c *TimedCache) StopCleanup() {
	close(c.stopChan)
}
