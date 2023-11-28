package my_cache

import (
	"sync"
)

// 对于Cache的再封装
type cache struct {
	mu         sync.Mutex
	lru        *Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//延迟初始化(Lazy Initialization)
	if c.lru == nil {
		c.lru = New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if get, o := c.lru.Get(key); o {
		//当初存进去的时候就是ByteView类型
		return get.(ByteView), o
	}
	return ByteView{}, false
}
