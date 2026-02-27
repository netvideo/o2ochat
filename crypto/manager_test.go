package crypto

import (
	"bytes"
	"testing"
	"time"
)

func TestNewCryptoManager(t *testing.T) {
	manager := NewCryptoManager(nil)
	if manager == nil {
		t.Fatal("Expected crypto manager to be created")
	}
}

func TestGenerateKeyPair(t *testing.T) {
	manager := NewCryptoManager(nil)

	keyPair, err := manager.GenerateKeyPair(AlgorithmEd25519)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if len(keyPair.PublicKey) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(keyPair.PublicKey))
	}

	if len(keyPair.PrivateKey) != 64 {
		t.Errorf("Expected private key length 64, got %d", len(keyPair.PrivateKey))
	}

	if keyPair.Algorithm != AlgorithmEd25519 {
		t.Errorf("Expected algorithm ed25519, got %s", keyPair.Algorithm)
	}

	if keyPair.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if keyPair.ExpiresAt.Before(keyPair.CreatedAt) {
		t.Error("Expected ExpiresAt to be after CreatedAt")
	}
}

func TestGenerateKeyPairX25519(t *testing.T) {
	manager := NewCryptoManager(nil)

	keyPair, err := manager.GenerateKeyPair(AlgorithmX25519)
	if err != nil {
		t.Fatalf("Failed to generate X25519 key pair: %v", err)
	}

	if len(keyPair.PublicKey) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(keyPair.PublicKey))
	}

	if len(keyPair.PrivateKey) != 32 {
		t.Errorf("Expected private key length 32, got %d", len(keyPair.PrivateKey))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	manager := NewCryptoManager(nil)

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("Hello, World!")

	config := DefaultEncryptionConfig()
	encrypted, err := manager.Encrypt(plaintext, key, config)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if len(encrypted.Ciphertext) == 0 {
		t.Error("Expected non-empty ciphertext")
	}

	if len(encrypted.Nonce) == 0 {
		t.Error("Expected non-empty nonce")
	}

	if len(encrypted.Tag) == 0 {
		t.Error("Expected non-empty tag")
	}

	decrypted, err := manager.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted text does not match original")
	}
}

func TestEncryptInvalidKeySize(t *testing.T) {
	manager := NewCryptoManager(nil)

	key := make([]byte, 16)
	plaintext := []byte("Hello")

	_, err := manager.Encrypt(plaintext, key, nil)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got: %v", err)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	manager := NewCryptoManager(nil)

	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	plaintext := []byte("Hello")
	encrypted, _ := manager.Encrypt(plaintext, key1, nil)

	_, err := manager.Decrypt(encrypted, key2)
	if err != ErrDecryptionFailed {
		t.Errorf("Expected ErrDecryptionFailed, got: %v", err)
	}
}

func TestSignVerify(t *testing.T) {
	manager := NewCryptoManager(nil)

	keyPair, _ := manager.GenerateKeyPair(AlgorithmEd25519)

	message := []byte("Test message")
	signed, err := manager.Sign(message, keyPair.PrivateKey)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if len(signed.Signature) != 64 {
		t.Errorf("Expected signature length 64, got %d", len(signed.Signature))
	}

	valid, err := manager.Verify(signed, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !valid {
		t.Error("Expected signature to be valid")
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	manager := NewCryptoManager(nil)

	_, err := manager.GenerateKeyPair(AlgorithmEd25519)
	if err != nil {
		t.Skip("Failed to generate key pair: ", err)
	}

	signed := &SignedMessage{
		Message:   []byte("test"),
		Signature: make([]byte, 64),
		Algorithm: AlgorithmEd25519,
	}

	valid, err := manager.Verify(signed, make([]byte, 32))
	if err != nil {
		t.Fatalf("Verify should not error: %v", err)
	}

	if valid {
		t.Error("Expected invalid signature")
	}
}

func TestHash(t *testing.T) {
	manager := NewCryptoManager(nil)

	data := []byte("test data")
	hash1, err := manager.Hash(data, AlgorithmSHA256)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if len(hash1) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash1))
	}

	hash2, _ := manager.Hash(data, AlgorithmSHA256)
	if !bytes.Equal(hash1, hash2) {
		t.Error("Expected same hash for same data")
	}

	hash3, _ := manager.Hash([]byte("different"), AlgorithmSHA256)
	if bytes.Equal(hash1, hash3) {
		t.Error("Expected different hash for different data")
	}
}

func TestRandomBytes(t *testing.T) {
	manager := NewCryptoManager(nil)

	bytes1, err := manager.RandomBytes(32)
	if err != nil {
		t.Fatalf("RandomBytes failed: %v", err)
	}

	if len(bytes1) != 32 {
		t.Errorf("Expected 32 bytes, got %d", len(bytes1))
	}

	bytes2, _ := manager.RandomBytes(32)
	if bytes.Equal(bytes1, bytes2) {
		t.Error("Expected different random bytes")
	}
}

func TestDeriveKey(t *testing.T) {
	manager := NewCryptoManager(nil)

	secret := []byte("shared secret")
	salt := []byte("random salt")
	info := []byte("context info")

	key, err := manager.DeriveKey(secret, salt, info, 32)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected derived key length 32, got %d", len(key))
	}

	key2, _ := manager.DeriveKey(secret, salt, info, 32)
	if !bytes.Equal(key, key2) {
		t.Error("Expected same derived key for same inputs")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	manager := NewCryptoManager(nil)

	a := []byte("test")
	b := []byte("test")
	c := []byte("different")

	if !manager.ConstantTimeCompare(a, b) {
		t.Error("Expected equal comparison")
	}

	if manager.ConstantTimeCompare(a, c) {
		t.Error("Expected not equal comparison")
	}
}

func TestSecureZeroMemory(t *testing.T) {
	manager := NewCryptoManager(nil)

	data := []byte("secret")
	manager.SecureZeroMemory(data)

	for _, b := range data {
		if b != 0 {
			t.Error("Expected memory to be zeroed")
		}
	}
}

func TestSecureRandomInt(t *testing.T) {
	manager := NewCryptoManager(nil)

	val, err := manager.SecureRandomInt(1, 100)
	if err != nil {
		t.Fatalf("SecureRandomInt failed: %v", err)
	}

	if val < 1 || val >= 100 {
		t.Errorf("Expected value in range [1, 100), got %d", val)
	}
}

func TestGenerateKeyID(t *testing.T) {
	manager := NewCryptoManager(nil)

	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	keyID := manager.GenerateKeyID(publicKey)
	if len(keyID) == 0 {
		t.Error("Expected non-empty key ID")
	}

	keyID2 := manager.GenerateKeyID(publicKey)
	if keyID != keyID2 {
		t.Error("Expected same key ID for same public key")
	}

	publicKey2 := make([]byte, 32)
	keyID3 := manager.GenerateKeyID(publicKey2)
	if keyID == keyID3 {
		t.Error("Expected different key ID for different public key")
	}
}

func TestCleanup(t *testing.T) {
	manager := NewCryptoManager(nil)

	err := manager.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
}

func TestDefaultEncryptionConfig(t *testing.T) {
	config := DefaultEncryptionConfig()

	if config.Algorithm != AlgorithmAESGCM {
		t.Errorf("Expected AES-GCM, got %s", config.Algorithm)
	}

	if config.KeySize != 32 {
		t.Errorf("Expected key size 32, got %d", config.KeySize)
	}

	if config.NonceSize != 12 {
		t.Errorf("Expected nonce size 12, got %d", config.NonceSize)
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if config.EncryptionAlgorithm != AlgorithmAESGCM {
		t.Errorf("Expected AES-GCM, got %s", config.EncryptionAlgorithm)
	}

	if config.SignatureAlgorithm != AlgorithmEd25519 {
		t.Errorf("Expected Ed25519, got %s", config.SignatureAlgorithm)
	}

	if config.KeyRotationInterval != 24*time.Hour {
		t.Error("Expected 24h key rotation")
	}
}
