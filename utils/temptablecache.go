package utils

import (
	"sync"
)

// TempTableCache is a thread-safe cache for storing temporary table mappings
var TempTableCache = &CacheSimple{
	Data:        make(map[string]interface{}),
	MaxSize:     100, // Note: this might need some tweaking if in a highly concurrent space...which is not the case now. Setting it too high risks too many temp tables..filling db.
	accessOrder: make([]string, 0),
}

// Defines a thread-safe in-memory cache
type CacheSimple struct {
	Data        map[string]interface{}
	MaxSize     int
	mu          sync.RWMutex
	accessOrder []string // Keeps track of insertion order for cleanup
}

// Retrieves a value from the cache
func (c *CacheSimple) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.Data[key]
	// If the key already exists, update its position in the queue, so it is kept a bit longer:
	if exists {
		c.moveToEnd(key)
	}
	return value, exists
}

// Adds a key-value pair to the cache
func (c *CacheSimple) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If the key already exists, update its value and move it to the end of the access order
	if _, exists := c.Data[key]; exists {
		c.Data[key] = value
		c.moveToEnd(key)
		return
	}

	// Add the new key-value pair
	c.Data[key] = value
	c.accessOrder = append(c.accessOrder, key)

	// If the cache exceeds its maximum size, remove the oldest / least recently accessed entry
	if len(c.accessOrder) > c.MaxSize {
		oldestKey := c.accessOrder[0]
		c.accessOrder = c.accessOrder[1:]
		delete(c.Data, oldestKey)
		c.removeFromOrder(oldestKey)
		// Optionally drop the corresponding table
		// dropTable(key) - probably not necessary if DB cleans up tmp tables automatically when session closes
	}
}

// Removes a key from the cache
func (c *CacheSimple) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.Data[key]; exists {
		delete(c.Data, key)
		c.removeFromOrder(key)
		// Optionally drop the corresponding table
		// dropTable(key) - probably not necessary if DB cleans up tmp tables automatically when session closes
	}
}

// Moves a key to the end of the access order
func (c *CacheSimple) moveToEnd(key string) {
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			c.accessOrder = append(c.accessOrder, key)
			break
		}
	}
}

// Removes a key from the access order
func (c *CacheSimple) removeFromOrder(key string) {
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
}
