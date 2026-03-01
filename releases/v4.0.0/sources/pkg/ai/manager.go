package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AIManager manages multiple AI providers
type AIManager struct {
	config           *AIManagerConfig
	providers        map[ProviderType]Translator
	activeProvider   ProviderType
	mu               sync.RWMutex
	translationCache *TranslationCache
}

// AIManagerConfig represents AI manager configuration
type AIManagerConfig struct {
	DefaultProvider ProviderType     `json:"default_provider"`
	Providers       []ProviderConfig `json:"providers"`
	EnableCache     bool             `json:"enable_cache"`
	CacheSize       int              `json:"cache_size"`
	CacheTTL        time.Duration    `json:"cache_ttl"`
	FallbackEnabled bool             `json:"fallback_enabled"`
	Timeout         time.Duration    `json:"timeout"`
}

// NewAIManager creates a new AI manager
func NewAIManager(config *AIManagerConfig) (*AIManager, error) {
	manager := &AIManager{
		config:    config,
		providers: make(map[ProviderType]Translator),
	}

	// Initialize providers
	for _, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			continue
		}

		var translator Translator
		switch providerConfig.Name {
		case ProviderOllama:
			translator = NewOllamaTranslator(&providerConfig)
		case ProviderOpenAI:
			translator = NewOpenAITranslator(&providerConfig)
		// Add more providers here
		default:
			return nil, fmt.Errorf("unknown provider: %s", providerConfig.Name)
		}

		manager.providers[providerConfig.Name] = translator
	}

	// Set active provider
	if config.DefaultProvider != "" {
		if _, exists := manager.providers[config.DefaultProvider]; exists {
			manager.activeProvider = config.DefaultProvider
		} else {
			return nil, fmt.Errorf("default provider not found: %s", config.DefaultProvider)
		}
	} else if len(manager.providers) > 0 {
		// Use first available provider
		for providerType := range manager.providers {
			manager.activeProvider = providerType
			break
		}
	}

	// Initialize cache
	if config.EnableCache {
		manager.translationCache = NewTranslationCache(config.CacheSize, config.CacheTTL)
	}

	return manager, nil
}

// Translate translates text using the active provider
func (m *AIManager) Translate(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	m.mu.RLock()
	provider := m.activeProvider
	m.mu.RUnlock()

	// Check cache
	if m.translationCache != nil {
		if cached, found := m.translationCache.Get(req); found {
			return cached, nil
		}
	}

	// Get provider
	translator, exists := m.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider not available: %s", provider)
	}

	// Translate
	resp, err := translator.Translate(ctx, req)
	if err != nil {
		// Try fallback if enabled
		if m.config.FallbackEnabled {
			resp, err = m.translateWithFallback(ctx, req, provider)
		}
		if err != nil {
			return nil, err
		}
	}

	// Cache result
	if m.translationCache != nil {
		m.translationCache.Set(req, resp)
	}

	return resp, nil
}

// translateWithFallback tries other providers if the active one fails
func (m *AIManager) translateWithFallback(ctx context.Context, req *TranslationRequest, skip ProviderType) (*TranslationResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for provider, translator := range m.providers {
		if provider == skip {
			continue
		}

		resp, err := translator.Translate(ctx, req)
		if err == nil {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("all providers failed")
}

// TranslateBatch translates multiple texts
func (m *AIManager) TranslateBatch(ctx context.Context, reqs []*TranslationRequest) ([]*TranslationResponse, error) {
	m.mu.RLock()
	provider := m.activeProvider
	m.mu.RUnlock()

	translator, exists := m.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider not available: %s", provider)
	}

	return translator.TranslateBatch(ctx, reqs)
}

// SetActiveProvider sets the active provider
func (m *AIManager) SetActiveProvider(provider ProviderType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[provider]; !exists {
		return fmt.Errorf("provider not available: %s", provider)
	}

	m.activeProvider = provider
	return nil
}

// GetActiveProvider returns the active provider
func (m *AIManager) GetActiveProvider() ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeProvider
}

// ListProviders returns list of available providers
func (m *AIManager) ListProviders() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(m.providers))
	for _, translator := range m.providers {
		infos = append(infos, translator.GetProviderInfo())
	}

	return infos
}

// GetProviderInfo returns info for a specific provider
func (m *AIManager) GetProviderInfo(provider ProviderType) (ProviderInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	translator, exists := m.providers[provider]
	if !exists {
		return ProviderInfo{}, fmt.Errorf("provider not found: %s", provider)
	}

	return translator.GetProviderInfo(), nil
}

// CheckHealth checks health of all providers
func (m *AIManager) CheckHealth(ctx context.Context) map[ProviderType]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[ProviderType]bool)

	for providerType, translator := range m.providers {
		// Check if translator has IsAvailable method
		if checker, ok := translator.(interface{ IsAvailable(context.Context) bool }); ok {
			health[providerType] = checker.IsAvailable(ctx)
		} else {
			health[providerType] = true // Assume healthy if no check method
		}
	}

	return health
}

// TranslationCache provides caching for translations
type TranslationCache struct {
	cache map[string]*cacheEntry
	size  int
	ttl   time.Duration
	mu    sync.RWMutex
}

type cacheEntry struct {
	response  *TranslationResponse
	expiresAt time.Time
}

// NewTranslationCache creates a new translation cache
func NewTranslationCache(size int, ttl time.Duration) *TranslationCache {
	return &TranslationCache{
		cache: make(map[string]*cacheEntry),
		size:  size,
		ttl:   ttl,
	}
}

// Get retrieves from cache
func (c *TranslationCache) Get(req *TranslationRequest) (*TranslationResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.buildCacheKey(req)
	entry, exists := c.cache[key]

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.response, true
}

// Set adds to cache
func (c *TranslationCache) Set(req *TranslationRequest, resp *TranslationResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean old entries if at capacity
	if len(c.cache) >= c.size {
		c.cleanOld()
	}

	key := c.buildCacheKey(req)
	c.cache[key] = &cacheEntry{
		response:  resp,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// buildCacheKey creates a cache key from request
func (c *TranslationCache) buildCacheKey(req *TranslationRequest) string {
	return fmt.Sprintf("%s:%s:%s", req.Text, req.SourceLang, req.TargetLang)
}

// cleanOld removes expired entries
func (c *TranslationCache) cleanOld() {
	now := time.Now()
	for key, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, key)
		}
	}
}
