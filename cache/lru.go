package cache

import "container/list"

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type entry struct {
	key   string
	value interface{}
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Put(key string, value interface{}) {

	if elem, found := c.items[key]; found {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if c.order.Len() >= c.capacity {
		c.evict()
	}

	e := &entry{key: key, value: value}
	elem := c.order.PushFront(e)
	c.items[key] = elem
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	elem, found := c.items[key]
	if !found {
		return nil, false
	}

	c.order.MoveToFront(elem)
	return elem.Value.(*entry).value, true
}

func (c *LRUCache) Delete(key string) {
	elem, found := c.items[key]
	if !found {
		return
	}
	c.removeElement(elem)
}

func (c *LRUCache) Clear() {
	c.items = make(map[string]*list.Element)
	c.order.Init()
}

func (c *LRUCache) Len() int {
	return c.order.Len()
}

func (c *LRUCache) evict() {
	oldest := c.order.Back()
	if oldest != nil {
		c.removeElement(oldest)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.order.Remove(elem)
	e := elem.Value.(*entry)
	delete(c.items, e.key)
}
