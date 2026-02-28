package decentralized

import (
	"sync"
	"time"
)

// RateLimiter implements IP-based rate limiting
type RateLimiter struct {
	requests   map[string][]time.Time // IP -> request timestamps
	limit      int                    // Max requests per window
	windowSize time.Duration          // Time window
	mu         sync.RWMutex
	stats      RateLimiterStats
	mu2        sync.RWMutex
}

// RateLimiterStats represents rate limiter statistics
type RateLimiterStats struct {
	TotalRequests   int `json:"total_requests"`
	AllowedRequests int `json:"allowed_requests"`
	BlockedRequests int `json:"blocked_requests"`
	UniqueIPs       int `json:"unique_ips"`
}

// RateLimiterConfig represents rate limiter configuration
type RateLimiterConfig struct {
	Limit      int           `json:"limit"`       // Max requests per window
	WindowSize time.Duration `json:"window_size"` // Time window
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		Limit:      100,
		WindowSize: time.Minute,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	return &RateLimiter{
		requests:   make(map[string][]time.Time),
		limit:      config.Limit,
		windowSize: config.WindowSize,
	}
}

// Allow checks if a request from IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.windowSize)

	// Get existing requests
	timestamps := rl.requests[ip]

	// Filter out old requests
	validTimestamps := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if limit exceeded
	if len(validTimestamps) >= rl.limit {
		rl.requests[ip] = validTimestamps

		rl.mu2.Lock()
		rl.stats.TotalRequests++
		rl.stats.BlockedRequests++
		rl.mu2.Unlock()

		return false
	}

	// Add new request
	validTimestamps = append(validTimestamps, now)
	rl.requests[ip] = validTimestamps

	rl.mu2.Lock()
	rl.stats.TotalRequests++
	rl.stats.AllowedRequests++
	rl.stats.UniqueIPs = len(rl.requests)
	rl.mu2.Unlock()

	return true
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() RateLimiterStats {
	rl.mu2.RLock()
	defer rl.mu2.RUnlock()
	return rl.stats
}

// Reset resets rate limiter state
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.requests = make(map[string][]time.Time)

	rl.mu2.Lock()
	rl.stats = RateLimiterStats{}
	rl.mu2.Unlock()
}

// Cleanup cleans up old entries
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.windowSize)

	for ip, timestamps := range rl.requests {
		validTimestamps := make([]time.Time, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		if len(validTimestamps) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validTimestamps
		}
	}

	rl.mu2.Lock()
	rl.stats.UniqueIPs = len(rl.requests)
	rl.mu2.Unlock()
}

// FrequencyLimiter implements connection frequency limiting
type FrequencyLimiter struct {
	connections map[string]time.Time // IP -> last connection time
	minInterval time.Duration        // Minimum interval between connections
	mu          sync.RWMutex
	stats       FrequencyLimiterStats
	mu2         sync.RWMutex
}

// FrequencyLimiterStats represents frequency limiter statistics
type FrequencyLimiterStats struct {
	TotalAttempts   int `json:"total_attempts"`
	AllowedAttempts int `json:"allowed_attempts"`
	BlockedAttempts int `json:"blocked_attempts"`
}

// FrequencyLimiterConfig represents frequency limiter configuration
type FrequencyLimiterConfig struct {
	MinInterval time.Duration `json:"min_interval"`
}

// DefaultFrequencyLimiterConfig returns default frequency limiter configuration
func DefaultFrequencyLimiterConfig() *FrequencyLimiterConfig {
	return &FrequencyLimiterConfig{
		MinInterval: time.Second,
	}
}

// NewFrequencyLimiter creates a new frequency limiter
func NewFrequencyLimiter(config *FrequencyLimiterConfig) *FrequencyLimiter {
	if config == nil {
		config = DefaultFrequencyLimiterConfig()
	}

	return &FrequencyLimiter{
		connections: make(map[string]time.Time),
		minInterval: config.MinInterval,
	}
}

// Allow checks if a connection from IP should be allowed
func (fl *FrequencyLimiter) Allow(ip string) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	now := time.Now()
	lastConnection, exists := fl.connections[ip]

	fl.mu2.Lock()
	fl.stats.TotalAttempts++
	fl.mu2.Unlock()

	if exists && now.Sub(lastConnection) < fl.minInterval {
		fl.mu2.Lock()
		fl.stats.BlockedAttempts++
		fl.mu2.Unlock()
		return false
	}

	fl.connections[ip] = now

	fl.mu2.Lock()
	fl.stats.AllowedAttempts++
	fl.mu2.Unlock()

	return true
}

// GetStats returns frequency limiter statistics
func (fl *FrequencyLimiter) GetStats() FrequencyLimiterStats {
	fl.mu2.RLock()
	defer fl.mu2.RUnlock()
	return fl.stats
}

// Reset resets frequency limiter state
func (fl *FrequencyLimiter) Reset() {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	fl.connections = make(map[string]time.Time)

	fl.mu2.Lock()
	fl.stats = FrequencyLimiterStats{}
	fl.mu2.Unlock()
}

// QuotaManager manages request quotas
type QuotaManager struct {
	quotas       map[string]*Quota // IP -> quota
	defaultQuota int
	mu           sync.RWMutex
	stats        QuotaManagerStats
	mu2          sync.RWMutex
}

// Quota represents a quota allocation
type Quota struct {
	Remaining int       `json:"remaining"`
	Limit     int       `json:"limit"`
	ResetAt   time.Time `json:"reset_at"`
	LastUsed  time.Time `json:"last_used"`
}

// QuotaManagerStats represents quota manager statistics
type QuotaManagerStats struct {
	TotalQuotas     int `json:"total_quotas"`
	ExhaustedQuotas int `json:"exhausted_quotas"`
	ActiveQuotas    int `json:"active_quotas"`
}

// QuotaManagerConfig represents quota manager configuration
type QuotaManagerConfig struct {
	DefaultQuota  int           `json:"default_quota"`
	ResetInterval time.Duration `json:"reset_interval"`
}

// DefaultQuotaManagerConfig returns default quota manager configuration
func DefaultQuotaManagerConfig() *QuotaManagerConfig {
	return &QuotaManagerConfig{
		DefaultQuota:  1000,
		ResetInterval: time.Hour,
	}
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager(config *QuotaManagerConfig) *QuotaManager {
	if config == nil {
		config = DefaultQuotaManagerConfig()
	}

	return &QuotaManager{
		quotas:       make(map[string]*Quota),
		defaultQuota: config.DefaultQuota,
	}
}

// Consume consumes quota for IP
func (qm *QuotaManager) Consume(ip string, amount int) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	now := time.Now()
	quota, exists := qm.quotas[ip]

	if !exists {
		// Create new quota
		quota = &Quota{
			Remaining: qm.defaultQuota - amount,
			Limit:     qm.defaultQuota,
			ResetAt:   now.Add(time.Hour),
			LastUsed:  now,
		}
		qm.quotas[ip] = quota

		qm.mu2.Lock()
		qm.stats.TotalQuotas++
		qm.stats.ActiveQuotas++
		qm.mu2.Unlock()

		return quota.Remaining >= 0
	}

	// Check if quota needs reset
	if now.After(quota.ResetAt) {
		quota.Remaining = quota.Limit
		quota.ResetAt = now.Add(time.Hour)
	}

	// Check if quota exhausted
	if quota.Remaining < amount {
		qm.mu2.Lock()
		qm.stats.ExhaustedQuotas++
		qm.mu2.Unlock()
		return false
	}

	// Consume quota
	quota.Remaining -= amount
	quota.LastUsed = now

	return true
}

// GetQuota gets quota for IP
func (qm *QuotaManager) GetQuota(ip string) *Quota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[ip]
	if !exists {
		return nil
	}

	return quota
}

// GetStats returns quota manager statistics
func (qm *QuotaManager) GetStats() QuotaManagerStats {
	qm.mu2.RLock()
	defer qm.mu2.RUnlock()
	return qm.stats
}

// Reset resets quota manager state
func (qm *QuotaManager) Reset() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.quotas = make(map[string]*Quota)

	qm.mu2.Lock()
	qm.stats = QuotaManagerStats{}
	qm.mu2.Unlock()
}
