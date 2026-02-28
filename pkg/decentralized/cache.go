package decentralized

import (
	"container/list"
	"sync"
	"time"
)

// DHTEnt represents a DHT cache entry
type DHTEnt struct {
	key       string
	value     interface{}
	expiresAt time.Time
	accessed  time.Time
	hits      int
	element   *list.Element // Reference to list element for O(1) removal
}

// DHTCache implements a DHT cache with LRU eviction
type DHTCache struct {
	capacity     int
	items        map[string]*DHTEnt
	evictionList *list.List
	mu           sync.RWMutex
	defaultTTL   time.Duration
}

// DHTCacheConfig represents DHT cache configuration
type DHTCacheConfig struct {
	Capacity   int           // Maximum number of entries
	DefaultTTL time.Duration // Default time-to-live for entries
}

// DefaultDHTCacheConfig returns default DHT cache configuration
func DefaultDHTCacheConfig() *DHTCacheConfig {
	return &DHTCacheConfig{
		Capacity:   10000, // 10,000 entries
		DefaultTTL: 5 * time.Minute,
	}
}

// NewDHTCache creates a new DHT cache
func NewDHTCache(config *DHTCacheConfig) *DHTCache {
	if config == nil {
		config = DefaultDHTCacheConfig()
	}

	return &DHTCache{
		capacity:     config.Capacity,
		items:        make(map[string]*DHTEnt),
		evictionList: list.New(),
		defaultTTL:   config.DefaultTTL,
	}
}

// Get retrieves a value from the cache
func (c *DHTCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ent, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(ent.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		defer c.mu.Unlock()
		// Double-check after acquiring write lock
		if ent, exists := c.items[key]; exists && time.Now().After(ent.expiresAt) {
			c.removeElement(ent)
		}
		return nil, false
	}

	// Update access time and hit count
	ent.accessed = time.Now()
	ent.hits++

	// Move to front of eviction list (most recently used)
	c.evictionList.MoveToFront(ent.element)

	return ent.value, true
}

// Set adds or updates a value in the cache
func (c *DHTCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds or updates a value with custom TTL
func (c *DHTCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Check if key exists
	if ent, exists := c.items[key]; exists {
		// Update existing entry
		ent.value = value
		ent.expiresAt = now.Add(ttl)
		ent.accessed = now
		ent.hits++
		if ent.element != nil {
			c.evictionList.MoveToFront(ent.element)
		}
		return
	}

	// Check capacity
	if len(c.items) >= c.capacity {
		// Evict least recently used
		c.evictOldest()
	}

	// Create new entry
	ent := &DHTEnt{
		key:       key,
		value:     value,
		expiresAt: now.Add(ttl),
		accessed:  now,
		hits:      1,
	}

	c.items[key] = ent
	c.evictionList.PushFront(ent)
}

// Delete removes a value from the cache
func (c *DHTCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, exists := c.items[key]; exists {
		c.removeElement(ent)
		return true
	}

	return false
}

// Clear clears all entries from the cache
func (c *DHTCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*DHTEnt)
	c.evictionList = list.New()
}

// Size returns the number of entries in the cache
func (c *DHTCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *DHTCache) Stats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalHits int
	var oldestAccess time.Time

	for _, ent := range c.items {
		totalHits += ent.hits
		if oldestAccess.IsZero() || ent.accessed.Before(oldestAccess) {
			oldestAccess = ent.accessed
		}
	}

	return &CacheStats{
		Size:      len(c.items),
		Capacity:  c.capacity,
		TotalHits: totalHits,
		AvgHits:   float64(totalHits) / float64(len(c.items)),
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size      int     // Current number of entries
	Capacity  int     // Maximum capacity
	TotalHits int     // Total number of hits
	AvgHits   float64 // Average hits per entry
}

// removeElement removes an element from the cache
func (c *DHTCache) removeElement(ent *DHTEnt) {
	delete(c.items, ent.key)
	if ent.element != nil {
		c.evictionList.Remove(ent.element)
	}
}

// evictOldest evicts the oldest entry
func (c *DHTCache) evictOldest() {
	elem := c.evictionList.Back()
	if elem != nil {
		ent := elem.Value.(*DHTEnt)
		c.removeElement(ent)
	}
}

// CleanupExpired removes all expired entries
func (c *DHTCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0

	for _, ent := range c.items {
		if now.After(ent.expiresAt) {
			c.removeElement(ent)
			count++
		}
	}

	return count
}

// StartCleanup starts periodic cleanup of expired entries
func (c *DHTCache) StartCleanup(interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.CleanupExpired()
			case <-stop:
				return
			}
		}
	}()

	return stop
}
