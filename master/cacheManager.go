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

// LFUCacheProvider implements Least-Frequently-Used cache.
// Ties in frequency are broken by insertion order (oldest evicted first).
type LFUCacheProvider struct {
	mu      sync.Mutex
	items   map[string]string
	freq    map[string]int      // key -> access count
	buckets map[int][]string    // freq -> ordered list of keys
	minFreq int
	cap     int
}

func NewLFUCacheProvider(capacity int) *LFUCacheProvider {
	return &LFUCacheProvider{
		items:   make(map[string]string),
		freq:    make(map[string]int),
		buckets: make(map[int][]string),
		cap:     capacity,
	}
}

// lfuPromote moves key from its current freq bucket to freq+1.
func (c *LFUCacheProvider) lfuPromote(key string) {
	f := c.freq[key]
	// remove from current bucket
	newBucket := make([]string, 0, len(c.buckets[f]))
	for _, k := range c.buckets[f] {
		if k != key {
			newBucket = append(newBucket, k)
		}
	}
	c.buckets[f] = newBucket
	if len(c.buckets[f]) == 0 {
		delete(c.buckets, f)
		if c.minFreq == f {
			c.minFreq = f + 1
		}
	}
	// add to next bucket
	c.freq[key] = f + 1
	c.buckets[f+1] = append(c.buckets[f+1], key)
}

func (c *LFUCacheProvider) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.items[key]
	if !ok {
		return "", false
	}
	c.lfuPromote(key)
	return val, true
}

func (c *LFUCacheProvider) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cap <= 0 {
		return
	}

	if _, exists := c.items[key]; exists {
		c.items[key] = value
		c.lfuPromote(key)
		return
	}

	// evict if at capacity
	if len(c.items) >= c.cap {
		bucket := c.buckets[c.minFreq]
		if len(bucket) > 0 {
			evict := bucket[0]
			c.buckets[c.minFreq] = bucket[1:]
			if len(c.buckets[c.minFreq]) == 0 {
				delete(c.buckets, c.minFreq)
			}
			delete(c.items, evict)
			delete(c.freq, evict)
		}
	}

	// insert new key with frequency 1
	c.items[key] = value
	c.freq[key] = 1
	c.buckets[1] = append(c.buckets[1], key)
	c.minFreq = 1
}

func (c *LFUCacheProvider) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.freq[key]
	if !ok {
		return
	}
	delete(c.items, key)
	delete(c.freq, key)
	newBucket := make([]string, 0, len(c.buckets[f]))
	for _, k := range c.buckets[f] {
		if k != key {
			newBucket = append(newBucket, k)
		}
	}
	c.buckets[f] = newBucket
	if len(c.buckets[f]) == 0 {
		delete(c.buckets, f)
	}
}

func (c *LFUCacheProvider) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// ARCCacheProvider implements Adaptive Replacement Cache.
// It maintains four lists: T1 (recent, seen once), T2 (frequent, seen 2+),
// B1 (ghost of T1), B2 (ghost of T2), and a target size p for T1.
type ARCCacheProvider struct {
	mu  sync.Mutex
	cap int
	p   int // target size for T1

	t1 []string          // recent cache (seen once)
	t2 []string          // frequent cache (seen 2+)
	b1 []string          // ghost of T1 (evicted from T1)
	b2 []string          // ghost of T2 (evicted from T2)

	cache map[string]string // actual stored values (t1 ∪ t2)
}

func NewARCCacheProvider(capacity int) *ARCCacheProvider {
	return &ARCCacheProvider{
		cap:   capacity,
		cache: make(map[string]string),
	}
}

func arcRemove(list []string, key string) []string {
	out := list[:0]
	for _, k := range list {
		if k != key {
			out = append(out, k)
		}
	}
	return out
}

func arcContains(list []string, key string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}
	return false
}

func (c *ARCCacheProvider) replace(inB2 bool) {
	t1Len := len(c.t1)
	if t1Len > 0 && (t1Len > c.p || (inB2 && t1Len == c.p)) {
		// evict from T1 to B1
		evict := c.t1[0]
		c.t1 = c.t1[1:]
		delete(c.cache, evict)
		c.b1 = append(c.b1, evict)
	} else if len(c.t2) > 0 {
		// evict from T2 to B2
		evict := c.t2[0]
		c.t2 = c.t2[1:]
		delete(c.cache, evict)
		c.b2 = append(c.b2, evict)
	}
}

func (c *ARCCacheProvider) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.cache[key]
	if !ok {
		return "", false
	}
	// move to MRU of T2
	if arcContains(c.t1, key) {
		c.t1 = arcRemove(c.t1, key)
		c.t2 = append(c.t2, key)
	} else {
		c.t2 = arcRemove(c.t2, key)
		c.t2 = append(c.t2, key)
	}
	return val, true
}

func (c *ARCCacheProvider) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cap <= 0 {
		return
	}

	// Case 1: already in cache (T1 or T2)
	if _, ok := c.cache[key]; ok {
		c.cache[key] = value
		if arcContains(c.t1, key) {
			c.t1 = arcRemove(c.t1, key)
			c.t2 = append(c.t2, key)
		} else {
			c.t2 = arcRemove(c.t2, key)
			c.t2 = append(c.t2, key)
		}
		return
	}

	// Case 2: in ghost B1
	if arcContains(c.b1, key) {
		delta := 1
		if len(c.b1) < len(c.b2) {
			delta = len(c.b2) / len(c.b1)
		}
		c.p = min(c.p+delta, c.cap)
		c.replace(false)
		c.b1 = arcRemove(c.b1, key)
		c.cache[key] = value
		c.t2 = append(c.t2, key)
		return
	}

	// Case 3: in ghost B2
	if arcContains(c.b2, key) {
		delta := 1
		if len(c.b2) < len(c.b1) {
			delta = len(c.b1) / len(c.b2)
		}
		c.p = max(c.p-delta, 0)
		c.replace(true)
		c.b2 = arcRemove(c.b2, key)
		c.cache[key] = value
		c.t2 = append(c.t2, key)
		return
	}

	// Case 4: new key
	total := len(c.t1) + len(c.t2)
	if total >= c.cap {
		if total == c.cap {
			c.replace(false)
		}
		// trim B1 if overflow
		if len(c.b1)+len(c.b2) >= c.cap {
			if len(c.b1) > 0 {
				c.b1 = c.b1[1:]
			} else if len(c.b2) > 0 {
				c.b2 = c.b2[1:]
			}
		}
	}
	c.cache[key] = value
	c.t1 = append(c.t1, key)
}

func (c *ARCCacheProvider) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
	c.t1 = arcRemove(c.t1, key)
	c.t2 = arcRemove(c.t2, key)
}

func (c *ARCCacheProvider) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cache)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
	case "lfu":
		log.Println("🔵 LFU Cache initialized with capacity:", capacity)
		return NewLFUCacheProvider(capacity)
	case "arc":
		log.Println("🔵 ARC Cache initialized with capacity:", capacity)
		return NewARCCacheProvider(capacity)
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
