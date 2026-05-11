package cache

import "sync"

type ARCCache struct {
	capacity int
	p        int

	T1 map[string]string
	T2 map[string]string
	B1 map[string]bool
	B2 map[string]bool

	mu sync.Mutex
}

func NewARCCache(capacity int) *ARCCache {
	return &ARCCache{
		capacity: capacity,
		T1:       make(map[string]string),
		T2:       make(map[string]string),
		B1:       make(map[string]bool),
		B2:       make(map[string]bool),
	}
}

func (c *ARCCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.T1[key]; ok {
		delete(c.T1, key)
		c.T2[key] = v
		return v, true
	}
	if v, ok := c.T2[key]; ok {
		return v, true
	}
	return "", false
}

func (c *ARCCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.T1[key]; ok {
		delete(c.T1, key)
		c.T2[key] = value
		return
	}

	if len(c.T1)+len(c.T2) >= c.capacity {
		for k := range c.T1 {
			delete(c.T1, k)
			c.B1[k] = true
			break
		}
	}

	c.T1[key] = value
}

func (c *ARCCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.T1, key)
	delete(c.T2, key)
}

func (c *ARCCache) Size() int {
	return len(c.T1) + len(c.T2)
}
