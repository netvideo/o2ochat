package storage

import (
	"container/list"
	"sync"
	"time"
)

// CacheItem represents a single cache entry with metadata.
type CacheItem struct {
	key       string
	value     []byte
	expiresAt time.Time
	accessed  time.Time
}

// LRUCacheManager implements an in-memory LRU (Least Recently Used) cache with TTL support.
type LRUCacheManager struct {
	mu          sync.RWMutex
	cache       map[string]*list.Element
	lruList     *list.List
	maxSize     int64
	currentSize int64
	stats       *CacheStats
	closed      bool
}

func NewLRUCacheManager(maxSizeBytes int) *LRUCacheManager {
	return &LRUCacheManager{
		cache:   make(map[string]*list.Element),
		lruList: list.New(),
		maxSize: int64(maxSizeBytes),
		stats: &CacheStats{
			Size:      maxSizeBytes,
			Items:     0,
			Hits:      0,
			Misses:    0,
			Evictions: 0,
		},
	}
}

func (c *LRUCacheManager) Set(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	if elem, exists := c.cache[key]; exists {
		c.lruList.MoveToFront(elem)
		item := elem.Value.(*CacheItem)
		c.currentSize -= int64(len(item.value))
		item.value = value
		item.accessed = time.Now()
		if ttl > 0 {
			item.expiresAt = time.Now().Add(ttl)
		} else {
			item.expiresAt = time.Time{}
		}
		c.currentSize += int64(len(value))
		return nil
	}

	item := &CacheItem{
		key:       key,
		value:     value,
		expiresAt: time.Time{},
		accessed:  time.Now(),
	}

	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}

	valueSize := int64(len(value))

	for c.currentSize+valueSize > c.maxSize && c.lruList.Len() > 0 {
		oldest := c.lruList.Back()
		if oldest != nil {
			oldItem := oldest.Value.(*CacheItem)
			c.currentSize -= int64(len(oldItem.value))
			c.lruList.Remove(oldest)
			delete(c.cache, oldItem.key)
			c.stats.Evictions++
		}
	}

	elem := c.lruList.PushFront(item)
	c.cache[key] = elem
	c.currentSize += valueSize
	c.stats.Items = c.lruList.Len()

	return nil
}

func (c *LRUCacheManager) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	elem, exists := c.cache[key]
	if !exists {
		c.stats.Misses++
		return nil, ErrCacheNotFound
	}

	item := elem.Value.(*CacheItem)

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.lruList.Remove(elem)
		delete(c.cache, key)
		c.currentSize -= int64(len(item.value))
		c.stats.Items = c.lruList.Len()
		c.stats.Misses++
		return nil, ErrCacheNotFound
	}

	c.lruList.MoveToFront(elem)
	item.accessed = time.Now()
	c.stats.Hits++

	value := make([]byte, len(item.value))
	copy(value, item.value)

	return value, nil
}

func (c *LRUCacheManager) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	elem, exists := c.cache[key]
	if !exists {
		return ErrCacheNotFound
	}

	item := elem.Value.(*CacheItem)
	c.lruList.Remove(elem)
	delete(c.cache, key)
	c.currentSize -= int64(len(item.value))
	c.stats.Items = c.lruList.Len()

	return nil
}

func (c *LRUCacheManager) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.currentSize = 0
	c.stats.Items = 0

	return nil
}

func (c *LRUCacheManager) GetCacheStats() (*CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrStorageNotInitialized
	}

	return &CacheStats{
		Size:      c.stats.Size,
		Items:     c.stats.Items,
		Hits:      c.stats.Hits,
		Misses:    c.stats.Misses,
		Evictions: c.stats.Evictions,
	}, nil
}

func (c *LRUCacheManager) Resize(newSize int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrStorageNotInitialized
	}

	c.maxSize = int64(newSize)
	c.stats.Size = newSize

	for c.currentSize > c.maxSize && c.lruList.Len() > 0 {
		oldest := c.lruList.Back()
		if oldest != nil {
			oldItem := oldest.Value.(*CacheItem)
			c.currentSize -= int64(len(oldItem.value))
			c.lruList.Remove(oldest)
			delete(c.cache, oldItem.key)
			c.stats.Evictions++
		}
	}

	c.stats.Items = c.lruList.Len()

	return nil
}

func (c *LRUCacheManager) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return nil
}

func (c *LRUCacheManager) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return 0
	}

	now := time.Now()
	cleaned := 0

	for elem := c.lruList.Back(); elem != nil; {
		item := elem.Value.(*CacheItem)
		prev := elem.Prev()

		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			c.currentSize -= int64(len(item.value))
			c.lruList.Remove(elem)
			delete(c.cache, item.key)
			cleaned++
			c.stats.Evictions++
		}

		elem = prev
	}

	c.stats.Items = c.lruList.Len()

	return cleaned
}
