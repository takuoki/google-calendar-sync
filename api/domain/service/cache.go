package service

import (
	"sync"
)

type Cache[K comparable, V any] struct {
	mu     sync.RWMutex
	values map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		values: make(map[K]V),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	return v, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}
