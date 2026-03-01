package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"sync"
	"time"
)

// KeyRotationPolicy represents key rotation policy
type KeyRotationPolicy struct {
	RotationInterval time.Duration `json:"rotation_interval"`
	MaxKeyAge        time.Duration `json:"max_key_age"`
	MinKeyAge        time.Duration `json:"min_key_age"`
	AutoRotate       bool          `json:"auto_rotate"`
}

// KeyInfo represents key metadata
type KeyInfo struct {
	KeyID      string    `json:"key_id"`
	Algorithm  string    `json:"algorithm"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	UsageCount int       `json:"usage_count"`
	IsActive   bool      `json:"is_active"`
	KeyType    string    `json:"key_type"` // "encryption", "signing", "authentication"
}

// KeyManager manages encryption keys with rotation
type KeyManager struct {
	keys       map[string]*KeyInfo
	currentKey string
	policy     *KeyRotationPolicy
	mu         sync.RWMutex
	stats      KeyManagerStats
	mu2        sync.RWMutex
}

// KeyManagerStats represents key manager statistics
type KeyManagerStats struct {
	TotalKeys     int `json:"total_keys"`
	ActiveKeys    int `json:"active_keys"`
	ExpiredKeys   int `json:"expired_keys"`
	RotationCount int `json:"rotation_count"`
}

// DefaultKeyRotationPolicy returns default key rotation policy
func DefaultKeyRotationPolicy() *KeyRotationPolicy {
	return &KeyRotationPolicy{
		RotationInterval: 24 * time.Hour,
		MaxKeyAge:        30 * 24 * time.Hour,
		MinKeyAge:        1 * time.Hour,
		AutoRotate:       true,
	}
}

// NewKeyManager creates a new key manager
func NewKeyManager(policy *KeyRotationPolicy) *KeyManager {
	if policy == nil {
		policy = DefaultKeyRotationPolicy()
	}

	km := &KeyManager{
		keys:   make(map[string]*KeyInfo),
		policy: policy,
	}

	// Create initial key
	km.generateKey("initial")

	return km
}

// GetCurrentKey gets the current active key
func (km *KeyManager) GetCurrentKey() *KeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	keyID := km.currentKey
	if keyID == "" {
		return nil
	}

	return km.keys[keyID]
}

// GenerateKey generates a new encryption key
func (km *KeyManager) generateKey(prefix string) string {
	keyID := prefix + "-" + time.Now().Format("20060102150405")

	now := time.Now()
	keyInfo := &KeyInfo{
		KeyID:      keyID,
		Algorithm:  "AES-256-GCM",
		CreatedAt:  now,
		ExpiresAt:  now.Add(km.policy.MaxKeyAge),
		LastUsedAt: now,
		UsageCount: 0,
		IsActive:   true,
		KeyType:    "encryption",
	}

	km.mu.Lock()
	km.keys[keyID] = keyInfo
	if km.currentKey == "" {
		km.currentKey = keyID
	}
	km.mu.Unlock()

	km.mu2.Lock()
	km.stats.TotalKeys++
	km.stats.ActiveKeys++
	km.mu2.Unlock()

	return keyID
}

// RotateKeys rotates encryption keys
func (km *KeyManager) RotateKeys() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	now := time.Now()

	// Check if rotation is allowed
	currentKey := km.keys[km.currentKey]
	if currentKey != nil && now.Sub(currentKey.CreatedAt) < km.policy.MinKeyAge {
		return errors.New("key rotation not allowed - minimum age not reached")
	}

	// Generate new key
	newKeyID := "rotated-" + time.Now().Format("20060102150405")
	keyInfo := &KeyInfo{
		KeyID:      newKeyID,
		Algorithm:  "AES-256-GCM",
		CreatedAt:  now,
		ExpiresAt:  now.Add(km.policy.MaxKeyAge),
		LastUsedAt: now,
		UsageCount: 0,
		IsActive:   true,
		KeyType:    "encryption",
	}

	// Deactivate old key
	if currentKey != nil {
		currentKey.IsActive = false
	}

	// Set new key
	km.keys[newKeyID] = keyInfo
	km.currentKey = newKeyID

	km.mu2.Lock()
	km.stats.TotalKeys++
	km.stats.RotationCount++
	km.mu2.Unlock()

	return nil
}

// GetKey gets a key by ID
func (km *KeyManager) GetKey(keyID string) *KeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	key, exists := km.keys[keyID]
	if !exists {
		return nil
	}

	return key
}

// GetActiveKeys gets all active keys
func (km *KeyManager) GetActiveKeys() []*KeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	activeKeys := make([]*KeyInfo, 0)
	for _, key := range km.keys {
		if key.IsActive && time.Now().Before(key.ExpiresAt) {
			activeKeys = append(activeKeys, key)
		}
	}

	return activeKeys
}

// ExpireKey expires a key
func (km *KeyManager) ExpireKey(keyID string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	key, exists := km.keys[keyID]
	if !exists {
		return errors.New("key not found")
	}

	key.IsActive = false
	key.ExpiresAt = time.Now()

	km.mu2.Lock()
	km.stats.ActiveKeys--
	km.stats.ExpiredKeys++
	km.mu2.Unlock()

	// If this was the current key, rotate
	if km.currentKey == keyID {
		km.mu.Unlock()
		err := km.RotateKeys()
		km.mu.Lock()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetStats returns key manager statistics
func (km *KeyManager) GetStats() KeyManagerStats {
	km.mu2.RLock()
	defer km.mu2.RUnlock()
	return km.stats
}

// CleanupExpiredKeys removes expired keys
func (km *KeyManager) CleanupExpiredKeys() int {
	km.mu.Lock()
	defer km.mu.Unlock()

	now := time.Now()
	removed := 0

	for keyID, key := range km.keys {
		if now.After(key.ExpiresAt) && !key.IsActive {
			delete(km.keys, keyID)
			removed++
		}
	}

	return removed
}

// AutoRotate starts automatic key rotation
func (km *KeyManager) AutoRotate(stopChan <-chan struct{}) {
	ticker := time.NewTicker(km.policy.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if km.policy.AutoRotate {
				km.RotateKeys()
			}
			km.CleanupExpiredKeys()
		case <-stopChan:
			return
		}
	}
}

// EncryptData encrypts data using current key
func EncryptData(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptData decrypts data using key
func DecryptData(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DeriveKey derives a key from password using PBKDF2
func DeriveKey(password, salt []byte, keyLen int) []byte {
	// Simple key derivation using SHA-256
	// In production, use proper PBKDF2 or Argon2
	key := make([]byte, keyLen)

	for i := 0; i < keyLen/32; i++ {
		hash := sha256.Sum256(append(password, salt...))
		copy(key[i*32:], hash[:])
		salt = hash[:]
	}

	return key
}

// GenerateSecureRandom generates secure random bytes
func GenerateSecureRandom(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
