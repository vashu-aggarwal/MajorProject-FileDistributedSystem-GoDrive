package master

import (
	"godrive/config"
	"log"
	"sync"
)

// CacheProvider is the interface all cache implementations must satisfy
type CacheProvider interface {
	Get(key string) (string, bool)
	Put(key string, value string)
	Delete(key string)
	Size() int
}

// InMemoryCache is a simple in-memory cache for frontend-selected algorithms
type InMemoryCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		store: make(map[string]string),
	}
}

func (c *InMemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

func (c *InMemoryCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// LRUCacheProvider wraps the LRUCache for interface compatibility
type LRUCacheProvider struct {
	mu    sync.Mutex
	items map[string]string
	queue []string
	cap   int
}

func NewLRUCacheProvider(capacity int) *LRUCacheProvider {
	return &LRUCacheProvider{
		items: make(map[string]string),
		queue: make([]string, 0),
		cap:   capacity,
	}
}

func (c *LRUCacheProvider) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.items[key]
	if !ok {
		return "", false
	}

	// Move to end (most recently used)
	newQueue := make([]string, 0)
	for _, k := range c.queue {
		if k != key {
			newQueue = append(newQueue, k)
		}
	}
	newQueue = append(newQueue, key)
	c.queue = newQueue

	return val, true
}

func (c *LRUCacheProvider) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		c.items[key] = value
		// Move to end
		newQueue := make([]string, 0)
		for _, k := range c.queue {
			if k != key {
				newQueue = append(newQueue, k)
			}
		}
		newQueue = append(newQueue, key)
		c.queue = newQueue
		return
	}

	if len(c.items) >= c.cap {
		// Evict least recently used (first in queue)
		if len(c.queue) > 0 {
			evict := c.queue[0]
			delete(c.items, evict)
			c.queue = c.queue[1:]
		}
	}

	c.items[key] = value
	c.queue = append(c.queue, key)
}

func (c *LRUCacheProvider) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	newQueue := make([]string, 0)
	for _, k := range c.queue {
		if k != key {
			newQueue = append(newQueue, k)
		}
	}
	c.queue = newQueue
}

func (c *LRUCacheProvider) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// FIFOCacheProvider implements First-In-First-Out cache
type FIFOCacheProvider struct {
	mu    sync.Mutex
	items map[string]string
	queue []string
	cap   int
}

func NewFIFOCacheProvider(capacity int) *FIFOCacheProvider {
	return &FIFOCacheProvider{
		items: make(map[string]string),
		queue: make([]string, 0),
		cap:   capacity,
	}
}

func (c *FIFOCacheProvider) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *FIFOCacheProvider) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		return // FIFO: ignore updates
	}

	if len(c.items) >= c.cap {
		if len(c.queue) > 0 {
			evict := c.queue[0]
			delete(c.items, evict)
			c.queue = c.queue[1:]
		}
	}

	c.items[key] = value
	c.queue = append(c.queue, key)
}

func (c *FIFOCacheProvider) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	newQueue := make([]string, 0)
	for _, k := range c.queue {
		if k != key {
			newQueue = append(newQueue, k)
		}
	}
	c.queue = newQueue
}

func (c *FIFOCacheProvider) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// RWCacheProvider is a simple Read-Write cache (no eviction, unlimited)
type RWCacheProvider struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewRWCacheProvider() *RWCacheProvider {
	return &RWCacheProvider{
		items: make(map[string]string),
	}
}

func (c *RWCacheProvider) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *RWCacheProvider) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *RWCacheProvider) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *RWCacheProvider) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// InitCache creates a cache instance based on algorithm name
func InitCache(algo string, capacity int) CacheProvider {
	switch algo {
	case "lru":
		log.Println("🔵 LRU Cache initialized with capacity:", capacity)
		return NewLRUCacheProvider(capacity)
	case "fifo":
		log.Println("🔵 FIFO Cache initialized with capacity:", capacity)
		return NewFIFOCacheProvider(capacity)
	case "memory":
		log.Println("🔵 In-Memory Cache initialized")
		return NewInMemoryCache()
	case "rw":
		log.Println("🔵 Read-Write Cache initialized (unlimited)")
		return NewRWCacheProvider()
	default:
		log.Println("⚠️  Unknown cache algorithm, defaulting to LRU")
		return NewLRUCacheProvider(capacity)
	}
}

// InitNodeSelector creates a node selector instance based on algorithm name
func InitNodeSelector(algo string, nodes []config.Node) NodeSelector {
	switch algo {
	case "roundRobin":
		log.Println("🔄 Round-Robin Node Selector initialized")
		return NewRoundRobinSelector(nodes)
	case "random":
		log.Println("🎲 Random Node Selector initialized")
		return NewRandomNodeSelector(nodes)
	case "leastNode":
		log.Println("⚖️  Least-Node (Load-Based) Selector initialized")
		return NewLeastNodeSelector(nodes)
	case "powerOfTwo":
		log.Println("🔱 Power-of-Two Node Selector initialized")
		return NewPowerOfTwoSelector(nodes)
	default:
		log.Println("⚠️  Unknown node selector algorithm, defaulting to Least-Node")
		return NewLeastNodeSelector(nodes)
	}
}
