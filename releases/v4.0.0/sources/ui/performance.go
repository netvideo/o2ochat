package ui

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Key        string
	Value      interface{}
	Expiration time.Time
	Size       int
}

type Cache struct {
	mu        sync.RWMutex
	items     map[string]*CacheEntry
	maxSize   int
	maxMemory int64
	currentMemory int64
	evictPolicy string
	hits      int
	misses    int
}

func NewCache(maxSize int, maxMemory int64, evictPolicy string) *Cache {
	return &Cache{
		items:       make(map[string]*CacheEntry),
		maxSize:    maxSize,
		maxMemory:  maxMemory,
		evictPolicy: evictPolicy,
	}
}

func (c *Cache) Set(key string, value interface{}, size int, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.items[key]; ok {
		c.currentMemory -= int64(existing.Size)
	}

	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		Size:       size,
		Expiration: time.Now().Add(ttl),
	}

	c.items[key] = entry
	c.currentMemory += int64(size)

	c.evictIfNeeded()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok {
		c.misses++
		return nil, false
	}

	if time.Now().After(entry.Expiration) {
		delete(c.items, key)
		c.currentMemory -= int64(entry.Size)
		c.misses++
		return nil, false
	}

	c.hits++
	return entry.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok {
		c.currentMemory -= int64(entry.Size)
		delete(c.items, key)
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheEntry)
	c.currentMemory = 0
}

func (c *Cache) evictIfNeeded() {
	for len(c.items) > c.maxSize || c.currentMemory > c.maxMemory {
		c.evict()
	}
}

func (c *Cache) evict() {
	if len(c.items) == 0 {
		return
	}

	var keyToDelete string

	switch c.evictPolicy {
	case "LRU":
		var oldest time.Time
		for key, entry := range c.items {
			if oldest.IsZero() || entry.Expiration.Before(oldest) {
				oldest = entry.Expiration
				keyToDelete = key
			}
		}
	case "LFU":
		// Simple LFU would need hit tracking
		for key := range c.items {
			keyToDelete = key
			break
		}
	default:
		for key := range c.items {
			keyToDelete = key
			break
		}
	}

	if entry, ok := c.items[keyToDelete]; ok {
		c.currentMemory -= int64(entry.Size)
		delete(c.items, keyToDelete)
	}
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}

func (c *Cache) Stats() (size int, memory int64, hits int, misses int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items), c.currentMemory, c.hits, c.misses
}

type MemoryPool struct {
	mu        sync.RWMutex
	pools     map[string]*sync.Pool
	allocator func() interface{}
}

func NewMemoryPool(allocator func() interface{}) *MemoryPool {
	return &MemoryPool{
		allocator: allocator,
		pools:     make(map[string]*sync.Pool),
	}
}

func (mp *MemoryPool) RegisterPool(name string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	alloc := mp.allocator
	mp.pools[name] = &sync.Pool{
		New: func() interface{} {
			return alloc()
		},
	}
}

func (mp *MemoryPool) Get(name string) interface{} {
	mp.mu.RLock()
	pool, ok := mp.pools[name]
	mp.mu.RUnlock()

	if !ok {
		return mp.allocator()
	}
	return pool.Get()
}

func (mp *MemoryPool) Put(name string, value interface{}) {
	mp.mu.RLock()
	pool, ok := mp.pools[name]
	mp.mu.RUnlock()

	if ok {
		pool.Put(value)
	}
}

type ObjectPool struct {
	mu       sync.RWMutex
	objects  chan interface{}
	factory  func() interface{}
	reset    func(interface{})
	size     int
}

func NewObjectPool(size int, factory func() interface{}, reset func(interface{})) *ObjectPool {
	pool := &ObjectPool{
		objects: make(chan interface{}, size),
		factory: factory,
		reset:   reset,
		size:    size,
	}

	for i := 0; i < size; i++ {
		pool.objects <- factory()
	}

	return pool
}

func (op *ObjectPool) Get() interface{} {
	select {
	case obj := <-op.objects:
		return obj
	default:
		return op.factory()
	}
}

func (op *ObjectPool) Put(obj interface{}) {
	if op.reset != nil {
		op.reset(obj)
	}

	select {
	case op.objects <- obj:
	default:
	}
}

func (op *ObjectPool) Len() int {
	return len(op.objects)
}

type ResourceManager struct {
	mu           sync.RWMutex
	resources    map[string]interface{}
	initializers map[string]func() interface{}
	finalizers   map[string]func(interface{})
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:    make(map[string]interface{}),
		initializers: make(map[string]func() interface{}),
		finalizers:   make(map[string]func(interface{})),
	}
}

func (rm *ResourceManager) Register(name string, init func() interface{}, final func(interface{})) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.initializers[name] = init
	rm.finalizers[name] = final
}

func (rm *ResourceManager) Acquire(name string) interface{} {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if r, ok := rm.resources[name]; ok {
		return r
	}

	if init, ok := rm.initializers[name]; ok {
		r := init()
		rm.resources[name] = r
		return r
	}

	return nil
}

func (rm *ResourceManager) Release(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if r, ok := rm.resources[name]; ok {
		if final, ok := rm.finalizers[name]; ok {
			final(r)
		}
		delete(rm.resources, name)
	}
}

func (rm *ResourceManager) ReleaseAll() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for name, r := range rm.resources {
		if final, ok := rm.finalizers[name]; ok {
			final(r)
		}
		delete(rm.resources, name)
	}
}

type Debouncer struct {
	mu      sync.Mutex
	timer   *time.Timer
	delay   time.Duration
	handler func()
}

func NewDebouncer(delay time.Duration, handler func()) *Debouncer {
	return &Debouncer{
		delay:   delay,
		handler: handler,
	}
}

func (d *Debouncer) Trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, func() {
		d.handler()
	})
}

func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
}

type Throttler struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{
		interval: interval,
	}
}

func (t *Throttler) Try() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastCall) >= t.interval {
		t.lastCall = now
		return true
	}
	return false
}

func (t *Throttler) Force() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastCall = time.Now()
}

type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	maxTokens float64
	refillRate float64
	lastRefill time.Time
}

func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}
