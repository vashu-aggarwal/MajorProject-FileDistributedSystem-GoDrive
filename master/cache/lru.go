package cache

import "sync"

type Node struct {
	key   string
	value string
	prev  *Node
	next  *Node
}

type LRUCache struct {
	capacity int
	items    map[string]*Node
	head     *Node
	tail     *Node
	mu       sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*Node),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.items[key]
	if !ok {
		return "", false
	}
	c.moveToFront(node)
	return node.value, true
}

func (c *LRUCache) Put(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, ok := c.items[key]; ok {
		node.value = value
		c.moveToFront(node)
		return
	}

	if len(c.items) >= c.capacity {
		delete(c.items, c.tail.key)
		c.removeNode(c.tail)
	}

	newNode := &Node{key: key, value: value}
	c.addToFront(newNode)
	c.items[key] = newNode
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, ok := c.items[key]; ok {
		c.removeNode(node)
		delete(c.items, key)
	}
}

func (c *LRUCache) Size() int {
	return len(c.items)
}

func (c *LRUCache) moveToFront(node *Node) {
	c.removeNode(node)
	c.addToFront(node)
}

func (c *LRUCache) addToFront(node *Node) {
	node.next = c.head
	node.prev = nil
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRUCache) removeNode(node *Node) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}
