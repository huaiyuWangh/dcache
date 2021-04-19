package local

import (
	"dcache/byteview"
	"sync"
)

type Cache struct {
	MaxBytes int64
	lru      *lru
	sync.Mutex
}

func (c *Cache) Set(key string, value byteview.ByteView) {
	c.Lock()
	defer c.Unlock()
	if c.lru == nil {
		c.lru = New(c.MaxBytes, nil)
	}
	c.lru.Set(key, value)
}

func (c *Cache) Get(key string) (value byteview.ByteView, ok bool) {
	c.Lock()
	defer c.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(byteview.ByteView), ok
	}
	return
}


