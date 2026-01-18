package cache

import "sync"

type LFUNode struct {
	key   string
	value string
	freq  int
}

type LFUCache struct {
	capacity int
	items    map[string]*LFUNode
	freqMap  map[int][]*LFUNode
	minFreq  int
	mu       sync.Mutex
}

func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		items:    make(map[string]*LFUNode),
		freqMap:  make(map[int][]*LFUNode),
	}
}

func (c *LFUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.items[key]
	if !ok {
		return "", false
	}
	c.increment(node)
	return node.value, true
}

func (c *LFUCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	if node, ok := c.items[key]; ok {
		node.value = value
		c.increment(node)
		return
	}

	if len(c.items) >= c.capacity {
		evict := c.freqMap[c.minFreq][0]
		c.freqMap[c.minFreq] = c.freqMap[c.minFreq][1:]
		delete(c.items, evict.key)
	}

	node := &LFUNode{key: key, value: value, freq: 1}
	c.items[key] = node
	c.freqMap[1] = append(c.freqMap[1], node)
	c.minFreq = 1
}

func (c *LFUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *LFUCache) Size() int {
	return len(c.items)
}

func (c *LFUCache) increment(node *LFUNode) {
	f := node.freq
	list := c.freqMap[f]
	for i, n := range list {
		if n == node {
			c.freqMap[f] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(c.freqMap[f]) == 0 && c.minFreq == f {
		c.minFreq++
	}
	node.freq++
	c.freqMap[node.freq] = append(c.freqMap[node.freq], node)
}
