package cache

import "sync"

type Node struct {
	Key   string
	Value string
	Prev  *Node
	Next  *Node
}

type LRUCache struct {
	Capacity int
	Items    map[string]*Node
	Head     *Node
	Tail     *Node
	Mu       sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		Capacity: capacity,
		Items:    make(map[string]*Node),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	node, ok := c.Items[key]
	if !ok {
		return "", false
	}
	c.moveToFront(node)
	return node.Value, true
}

func (c *LRUCache) Put(key string, value string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if node, ok := c.Items[key]; ok {
		node.Value = value
		c.moveToFront(node)
		return
	}

	if len(c.Items) >= c.Capacity {
		delete(c.Items, c.Tail.Key)
		c.removeNode(c.Tail)
	}

	newNode := &Node{Key: key, Value: value}
	c.addToFront(newNode)
	c.Items[key] = newNode
}

func (c *LRUCache) Delete(key string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if node, ok := c.Items[key]; ok {
		c.removeNode(node)
		delete(c.Items, key)
	}
}

func (c *LRUCache) Size() int {
	return len(c.Items)
}

func (c *LRUCache) moveToFront(node *Node) {
	c.removeNode(node)
	c.addToFront(node)
}

func (c *LRUCache) addToFront(node *Node) {
	node.Next = c.Head
	node.Prev = nil
	if c.Head != nil {
		c.Head.Prev = node
	}
	c.Head = node
	if c.Tail == nil {
		c.Tail = node
	}
}

func (c *LRUCache) removeNode(node *Node) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		c.Head = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		c.Tail = node.Prev
	}
}
