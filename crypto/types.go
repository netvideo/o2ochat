package crypto

import (
	"time"
)

type AlgorithmType string

const (
	AlgorithmAESGCM   AlgorithmType = "aes-gcm"
	AlgorithmChaCha20 AlgorithmType = "chacha20-poly1305"
	AlgorithmEd25519  AlgorithmType = "ed25519"
	AlgorithmX25519   AlgorithmType = "x25519"
	AlgorithmSHA256   AlgorithmType = "sha256"
	AlgorithmSHA3_256 AlgorithmType = "sha3-256"
	AlgorithmBLAKE2b  AlgorithmType = "blake2b"
)

type KeyPair struct {
	PublicKey  []byte        `json:"public_key"`
	PrivateKey []byte        `json:"private_key"`
	Algorithm  AlgorithmType `json:"algorithm"`
	CreatedAt  time.Time     `json:"created_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
}

type EncryptionConfig struct {
	Algorithm  AlgorithmType `json:"algorithm"`
	KeySize    int           `json:"key_size"`
	NonceSize  int           `json:"nonce_size"`
	TagSize    int           `json:"tag_size"`
	UseHKDF    bool          `json:"use_hkdf"`
	SaltSize   int           `json:"salt_size"`
}

type EncryptedMessage struct {
	Ciphertext []byte        `json:"ciphertext"`
	Nonce      []byte        `json:"nonce"`
	Tag        []byte        `json:"tag"`
	Algorithm  AlgorithmType `json:"algorithm"`
	Version    string        `json:"version"`
	Timestamp  time.Time     `json:"timestamp"`
}

type SignedMessage struct {
	Message   []byte        `json:"message"`
	Signature []byte        `json:"signature"`
	PublicKey []byte        `json:"public_key"`
	Algorithm AlgorithmType `json:"algorithm"`
}

type KeyExchangeResult struct {
	SharedSecret []byte    `json:"shared_secret"`
	PublicKey    []byte    `json:"public_key"`
	SessionID    string    `json:"session_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type SessionKeys struct {
	EncryptionKey []byte `json:"encryption_key"`
	AuthTag       []byte `json:"auth_tag"`
	SessionID     string `json:"session_id"`
}

type SecurityConfig struct {
	EncryptionAlgorithm   AlgorithmType `json:"encryption_algorithm"`
	EncryptionKeySize     int            `json:"encryption_key_size"`
	SignatureAlgorithm    AlgorithmType `json:"signature_algorithm"`
	KeyExchangeAlgorithm  AlgorithmType `json:"key_exchange_algorithm"`
	KeyRotationInterval   time.Duration `json:"key_rotation_interval"`
	HashAlgorithm         AlgorithmType `json:"hash_algorithm"`
	MinEntropyBits        int            `json:"min_entropy_bits"`
	MinPasswordLength     int            `json:"min_password_length"`
	MaxFailedAttempts     int            `json:"max_failed_attempts"`
	LockoutDuration       time.Duration  `json:"lockout_duration"`
}

func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		Algorithm: AlgorithmAESGCM,
		KeySize:   32,
		NonceSize: 12,
		TagSize:   16,
		UseHKDF:   true,
		SaltSize:  32,
	}
}

func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EncryptionAlgorithm:  AlgorithmAESGCM,
		EncryptionKeySize:    32,
		SignatureAlgorithm:   AlgorithmEd25519,
		KeyExchangeAlgorithm: AlgorithmX25519,
		KeyRotationInterval:  24 * time.Hour,
		HashAlgorithm:         AlgorithmSHA256,
		MinEntropyBits:        256,
		MinPasswordLength:     12,
		MaxFailedAttempts:     5,
		LockoutDuration:       15 * time.Minute,
	}
}
