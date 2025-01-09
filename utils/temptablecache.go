package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// TempTableCache is a thread-safe cache for storing temporary table mappings
var TempTableCache = &Cache{
	data: sync.Map{},
}

// Defines a thread-safe in-memory cache
type Cache struct {
	data sync.Map
}

// Retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.data.Load(key)
}

// Adds a key-value pair to the cache
func (c *Cache) Set(key string, value interface{}) {
	c.data.Store(key, value)

}

// Removes a key from the cache
func (c *Cache) Delete(key string) {
	c.data.Delete(key)
}

// Creates a unique hash from a given string
func GenerateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:]) // Convert to hexadecimal string
}
