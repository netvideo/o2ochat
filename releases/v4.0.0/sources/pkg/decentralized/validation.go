package decentralized

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
	"time"
)

// PeerValidator validates peer IDs
type PeerValidator struct {
	blacklist map[string]time.Time // NodeID -> ban expiry
	whitelist map[string]bool      // Trusted NodeIDs
	validated map[string]bool      // Previously validated NodeIDs
	mu        sync.RWMutex
	stats     PeerValidatorStats
	mu2       sync.RWMutex
}

// PeerValidatorStats represents validator statistics
type PeerValidatorStats struct {
	TotalValidations int `json:"total_validations"`
	ValidPeerIDs     int `json:"valid_peer_ids"`
	InvalidPeerIDs   int `json:"invalid_peer_ids"`
	BlacklistedPeers int `json:"blacklisted_peers"`
	WhitelistedPeers int `json:"whitelisted_peers"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	IsValid   bool      `json:"is_valid"`
	Errors    []string  `json:"errors"`
	Warning   string    `json:"warning"`
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
}

// PeerValidatorConfig represents validator configuration
type PeerValidatorConfig struct {
	MinLength      int  `json:"min_length"`
	MaxLength      int  `json:"max_length"`
	RequireBase58  bool `json:"require_base58"`
	RequireHex     bool `json:"require_hex"`
	CheckBlacklist bool `json:"check_blacklist"`
	CheckWhitelist bool `json:"check_whitelist"`
}

// DefaultPeerValidatorConfig returns default validator configuration
func DefaultPeerValidatorConfig() *PeerValidatorConfig {
	return &PeerValidatorConfig{
		MinLength:      20,
		MaxLength:      64,
		RequireBase58:  false,
		RequireHex:     false,
		CheckBlacklist: true,
		CheckWhitelist: false,
	}
}

// NewPeerValidator creates a new peer validator
func NewPeerValidator(config *PeerValidatorConfig) *PeerValidator {
	if config == nil {
		config = DefaultPeerValidatorConfig()
	}

	return &PeerValidator{
		blacklist: make(map[string]time.Time),
		whitelist: make(map[string]bool),
		validated: make(map[string]bool),
	}
}

// Validate validates a peer ID
func (pv *PeerValidator) Validate(nodeID string) *ValidationResult {
	pv.mu2.Lock()
	pv.stats.TotalValidations++
	pv.mu2.Unlock()

	result := &ValidationResult{
		IsValid:   true,
		Errors:    make([]string, 0),
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}

	// Check length
	if len(nodeID) < 20 || len(nodeID) > 64 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid length (must be 20-64 characters)")
	}

	// Check format (Base58 or Hex)
	isBase58 := isValidBase58(nodeID)
	isHex := isValidHex(nodeID)

	if !isBase58 && !isHex {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid format (must be Base58 or Hex)")
		result.Warning = "Peer ID should be Base58 or Hex encoded"
	}

	// Check blacklist
	pv.mu.RLock()
	if banExpiry, exists := pv.blacklist[nodeID]; exists {
		if time.Now().Before(banExpiry) {
			pv.mu.RUnlock()
			result.IsValid = false
			result.Errors = append(result.Errors, "Peer ID is blacklisted")
			pv.mu2.Lock()
			pv.stats.BlacklistedPeers++
			pv.mu2.Unlock()
			return result
		}
		// Ban expired, remove
		delete(pv.blacklist, nodeID)
		pv.mu.RUnlock()
	} else {
		pv.mu.RUnlock()
	}

	// Check whitelist (if enabled)
	pv.mu.RLock()
	if pv.whitelist[nodeID] {
		pv.mu.RUnlock()
		pv.mu2.Lock()
		pv.stats.WhitelistedPeers++
		pv.mu2.Unlock()
		result.IsValid = true
		result.Warning = "Trusted peer (whitelisted)"
		return result
	}
	pv.mu.RUnlock()

	// Update stats
	if result.IsValid {
		pv.mu2.Lock()
		pv.stats.ValidPeerIDs++
		pv.mu2.Unlock()

		// Cache validated peer
		pv.mu.Lock()
		pv.validated[nodeID] = true
		pv.mu.Unlock()
	} else {
		pv.mu2.Lock()
		pv.stats.InvalidPeerIDs++
		pv.mu2.Unlock()
	}

	return result
}

// Blacklist adds a peer ID to the blacklist
func (pv *PeerValidator) Blacklist(nodeID string, duration time.Duration) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	expiry := time.Now().Add(duration)
	pv.blacklist[nodeID] = expiry
}

// Whitelist adds a peer ID to the whitelist
func (pv *PeerValidator) Whitelist(nodeID string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	pv.whitelist[nodeID] = true
}

// RemoveBlacklist removes a peer ID from the blacklist
func (pv *PeerValidator) RemoveBlacklist(nodeID string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	delete(pv.blacklist, nodeID)
}

// RemoveWhitelist removes a peer ID from the whitelist
func (pv *PeerValidator) RemoveWhitelist(nodeID string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	delete(pv.whitelist, nodeID)
}

// IsBlacklisted checks if a peer ID is blacklisted
func (pv *PeerValidator) IsBlacklisted(nodeID string) bool {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	if banExpiry, exists := pv.blacklist[nodeID]; exists {
		return time.Now().Before(banExpiry)
	}
	return false
}

// IsWhitelisted checks if a peer ID is whitelisted
func (pv *PeerValidator) IsWhitelisted(nodeID string) bool {
	pv.mu.RLock()
	defer pv.mu.RUnlock()
	return pv.whitelist[nodeID]
}

// GetStats returns validator statistics
func (pv *PeerValidator) GetStats() PeerValidatorStats {
	pv.mu2.RLock()
	defer pv.mu2.RUnlock()
	return pv.stats
}

// ClearCache clears the validation cache
func (pv *PeerValidator) ClearCache() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.validated = make(map[string]bool)
}

// Helper function to validate Base58
func isValidBase58(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Base58 alphabet (no 0, O, I, l)
	base58Regex := regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]+$`)
	return base58Regex.MatchString(s)
}

// Helper function to validate Hex
func isValidHex(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Hex alphabet (0-9, a-f, A-F)
	hexRegex := regexp.MustCompile(`^[0-9a-fA-F]+$`)
	return hexRegex.MatchString(s)
}

// GenerateNodeID generates a node ID from public key
func GenerateNodeID(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// ValidateNodeID validates a node ID format
func ValidateNodeID(nodeID string) bool {
	// Check length
	if len(nodeID) < 20 || len(nodeID) > 64 {
		return false
	}

	// Check format
	return isValidBase58(nodeID) || isValidHex(nodeID)
}

// SanitizeNodeID sanitizes a node ID
func SanitizeNodeID(nodeID string) string {
	// Trim whitespace
	nodeID = strings.TrimSpace(nodeID)

	// Convert to lowercase for hex
	if isValidHex(nodeID) {
		return strings.ToLower(nodeID)
	}

	return nodeID
}

// CompareNodeIDs compares two node IDs
func CompareNodeIDs(id1, id2 string) bool {
	// Sanitize both IDs
	id1 = SanitizeNodeID(id1)
	id2 = SanitizeNodeID(id2)

	// Direct comparison
	if id1 == id2 {
		return true
	}

	// Try hex comparison
	if isValidHex(id1) && isValidHex(id2) {
		return strings.ToLower(id1) == strings.ToLower(id2)
	}

	return false
}
