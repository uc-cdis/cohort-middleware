package utils

import (
	"sync"
)

// TempTableCache is a thread-safe cache for storing temporary table mappings
var TempTableCache = &CacheSimple{
	data:        make(map[string]interface{}),
	maxSize:     10,
	accessOrder: make([]string, 0),
}

// Defines a thread-safe in-memory cache
type CacheSimple struct {
	data        map[string]interface{}
	mu          sync.RWMutex
	maxSize     int
	accessOrder []string // Keeps track of insertion order for cleanup
}

// Retrieves a value from the cache
func (c *CacheSimple) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
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
	if _, exists := c.data[key]; exists {
		c.data[key] = value
		c.moveToEnd(key)
		return
	}

	// Add the new key-value pair
	c.data[key] = value
	c.accessOrder = append(c.accessOrder, key)

	// If the cache exceeds its maximum size, remove the oldest / least recently accessed entry
	if len(c.accessOrder) > c.maxSize {
		oldestKey := c.accessOrder[0]
		c.accessOrder = c.accessOrder[1:]
		delete(c.data, oldestKey)

		// Optionally drop the corresponding table
		dropTable(oldestKey)
	}
}

// Removes a key from the cache
func (c *CacheSimple) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.removeFromOrder(key)
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

// Drops a temporary table (to be implemented based on your database logic)
func dropTable(key string) {
	// Add logic to drop the temporary table associated with the key
	// Example: db.Exec(fmt.Sprintf("DROP TABLE %s", key))
}
