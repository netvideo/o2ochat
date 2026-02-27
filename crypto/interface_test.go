package crypto

import (
	"testing"
	"time"
)

func TestAlgorithmTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected AlgorithmType
	}{
		{"AESGCM", AlgorithmAESGCM},
		{"ChaCha20", AlgorithmChaCha20},
		{"Ed25519", AlgorithmEd25519},
		{"X25519", AlgorithmX25519},
		{"SHA256", AlgorithmSHA256},
		{"SHA3_256", AlgorithmSHA3_256},
		{"BLAKE2b", AlgorithmBLAKE2b},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("algorithm type should not be empty")
			}
		})
	}
}

func TestKeyPairCreation(t *testing.T) {
	kp := &KeyPair{
		PublicKey:  []byte("public-key-test"),
		PrivateKey: []byte("private-key-test"),
		Algorithm:  AlgorithmEd25519,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	if kp.PublicKey == nil {
		t.Error("public key should not be nil")
	}
	if kp.PrivateKey == nil {
		t.Error("private key should not be nil")
	}
	if kp.Algorithm != AlgorithmEd25519 {
		t.Errorf("expected algorithm Ed25519, got %v", kp.Algorithm)
	}
}

func TestEncryptionConfig(t *testing.T) {
	config := DefaultEncryptionConfig()

	if config.Algorithm != AlgorithmAESGCM {
		t.Errorf("expected AESGCM, got %v", config.Algorithm)
	}
	if config.KeySize != 32 {
		t.Errorf("expected key size 32, got %d", config.KeySize)
	}
	if config.NonceSize != 12 {
		t.Errorf("expected nonce size 12, got %d", config.NonceSize)
	}
	if config.TagSize != 16 {
		t.Errorf("expected tag size 16, got %d", config.TagSize)
	}
	if !config.UseHKDF {
		t.Error("HKDF should be enabled by default")
	}
	if config.SaltSize != 32 {
		t.Errorf("expected salt size 32, got %d", config.SaltSize)
	}
}

func TestEncryptedMessage(t *testing.T) {
	msg := &EncryptedMessage{
		Ciphertext: []byte("ciphertext-data"),
		Nonce:      []byte("123456789012"),
		Tag:        []byte("authentication-tag"),
		Algorithm:  AlgorithmAESGCM,
		Version:    "1.0",
		Timestamp:  time.Now(),
	}

	if msg.Ciphertext == nil {
		t.Error("ciphertext should not be nil")
	}
	if msg.Nonce == nil {
		t.Error("nonce should not be nil")
	}
	if msg.Tag == nil {
		t.Error("tag should not be nil")
	}
	if msg.Algorithm != AlgorithmAESGCM {
		t.Errorf("expected AESGCM, got %v", msg.Algorithm)
	}
}

func TestSignedMessage(t *testing.T) {
	signed := &SignedMessage{
		Message:   []byte("message-content"),
		Signature: []byte("signature-data"),
		PublicKey: []byte("public-key"),
		Algorithm: AlgorithmEd25519,
	}

	if signed.Message == nil {
		t.Error("message should not be nil")
	}
	if signed.Signature == nil {
		t.Error("signature should not be nil")
	}
	if signed.Algorithm != AlgorithmEd25519 {
		t.Errorf("expected Ed25519, got %v", signed.Algorithm)
	}
}

func TestKeyExchangeResult(t *testing.T) {
	result := &KeyExchangeResult{
		SharedSecret: []byte("shared-secret"),
		PublicKey:    []byte("public-key"),
		SessionID:    "session-123",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if result.SharedSecret == nil {
		t.Error("shared secret should not be nil")
	}
	if result.PublicKey == nil {
		t.Error("public key should not be nil")
	}
	if result.SessionID == "" {
		t.Error("session ID should not be empty")
	}
}

func TestSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if config.EncryptionAlgorithm != AlgorithmAESGCM {
		t.Errorf("expected AESGCM, got %v", config.EncryptionAlgorithm)
	}
	if config.EncryptionKeySize != 32 {
		t.Errorf("expected key size 32, got %d", config.EncryptionKeySize)
	}
	if config.SignatureAlgorithm != AlgorithmEd25519 {
		t.Errorf("expected Ed25519, got %v", config.SignatureAlgorithm)
	}
	if config.KeyExchangeAlgorithm != AlgorithmX25519 {
		t.Errorf("expected X25519, got %v", config.KeyExchangeAlgorithm)
	}
	if config.HashAlgorithm != AlgorithmSHA256 {
		t.Errorf("expected SHA256, got %v", config.HashAlgorithm)
	}
	if config.MinEntropyBits != 256 {
		t.Errorf("expected 256 entropy bits, got %d", config.MinEntropyBits)
	}
	if config.MinPasswordLength != 12 {
		t.Errorf("expected 12 min password length, got %d", config.MinPasswordLength)
	}
	if config.MaxFailedAttempts != 5 {
		t.Errorf("expected 5 max failed attempts, got %d", config.MaxFailedAttempts)
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrInvalidKeySize, "ErrInvalidKeySize"},
		{ErrInvalidNonceSize, "ErrInvalidNonceSize"},
		{ErrEncryptionFailed, "ErrEncryptionFailed"},
		{ErrDecryptionFailed, "ErrDecryptionFailed"},
		{ErrSignatureFailed, "ErrSignatureFailed"},
		{ErrVerificationFailed, "ErrVerificationFailed"},
		{ErrKeyGenerationFailed, "ErrKeyGenerationFailed"},
		{ErrKeyExchangeFailed, "ErrKeyExchangeFailed"},
		{ErrKeyNotFound, "ErrKeyNotFound"},
		{ErrKeyAlreadyExists, "ErrKeyAlreadyExists"},
		{ErrKeyExpired, "ErrKeyExpired"},
		{ErrInvalidAlgorithm, "ErrInvalidAlgorithm"},
		{ErrInvalidPublicKey, "ErrInvalidPublicKey"},
		{ErrInvalidPrivateKey, "ErrInvalidPrivateKey"},
		{ErrRandomGeneration, "ErrRandomGeneration"},
		{ErrKeyDerivationFailed, "ErrKeyDerivationFailed"},
		{ErrSessionNotFound, "ErrSessionNotFound"},
		{ErrSessionExpired, "ErrSessionExpired"},
		{ErrPasswordMismatch, "ErrPasswordMismatch"},
		{ErrPasswordTooShort, "ErrPasswordTooShort"},
		{ErrInvalidHash, "ErrInvalidHash"},
		{ErrMemoryCleanup, "ErrMemoryCleanup"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestCryptoError(t *testing.T) {
	innerErr := ErrKeyGenerationFailed
	cryptoErr := NewCryptoError("KEY_GEN", "key generation failed", innerErr)

	if cryptoErr.Code != "KEY_GEN" {
		t.Errorf("expected code KEY_GEN, got %s", cryptoErr.Code)
	}
	if cryptoErr.Message != "key generation failed" {
		t.Errorf("expected message 'key generation failed', got %s", cryptoErr.Message)
	}
	if cryptoErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if cryptoErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ CryptoManager = nil
	var _ KeyExchange = nil
	var _ KeyStorage = nil
	var _ CryptoUtil = nil
	var _ ForwardSecrecyProtocol = nil
}
