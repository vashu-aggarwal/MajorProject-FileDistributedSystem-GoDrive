package cache

import "sync"

type FIFOCache struct {
	capacity int
	queue    []string
	items    map[string]string
	mu       sync.Mutex
}

func NewFIFOCache(capacity int) *FIFOCache {
	return &FIFOCache{
		capacity: capacity,
		queue:    []string{},
		items:    make(map[string]string),
	}
}

func (c *FIFOCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *FIFOCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		return
	}

	if len(c.queue) >= c.capacity {
		evict := c.queue[0]
		c.queue = c.queue[1:]
		delete(c.items, evict)
	}

	c.queue = append(c.queue, key)
	c.items[key] = value
}

func (c *FIFOCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	for i, k := range c.queue {
		if k == key {
			c.queue = append(c.queue[:i], c.queue[i+1:]...)
			break
		}
	}
}

func (c *FIFOCache) Size() int {
	return len(c.items)
}
