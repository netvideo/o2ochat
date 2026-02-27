package identity

import (
	"testing"
	"time"
)

func TestKeyTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected KeyType
	}{
		{"Ed25519", KeyTypeEd25519},
		{"RSA", KeyTypeRSA},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("key type should not be empty")
			}
		})
	}
}

func TestPeerIDEncodings(t *testing.T) {
	tests := []struct {
		name     string
		expected PeerIDEncoding
	}{
		{"Base58", PeerIDEncodingBase58},
		{"Hex", PeerIDEncodingHex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("encoding should not be empty")
			}
		})
	}
}

func TestIdentityCreation(t *testing.T) {
	identity := &Identity{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("public-key-test"),
		PrivateKey: []byte("private-key-test"),
		CreatedAt:  time.Now(),
	}

	if identity.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
	if identity.PublicKey == nil {
		t.Error("public key should not be nil")
	}
	if identity.PrivateKey == nil {
		t.Error("private key should not be nil")
	}
}

func TestIdentityConfig(t *testing.T) {
	config := DefaultIdentityConfig()

	if config.KeyType != KeyTypeEd25519 {
		t.Errorf("expected Ed25519, got %v", config.KeyType)
	}
	if config.KeyLength != 256 {
		t.Errorf("expected key length 256, got %d", config.KeyLength)
	}
	if config.PeerIDEncoding != PeerIDEncodingBase58 {
		t.Errorf("expected Base58, got %v", config.PeerIDEncoding)
	}
}

func TestIdentityMetadata(t *testing.T) {
	metadata := &IdentityMetadata{
		PeerID:      "test-peer-id",
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		DisplayName: "Test User",
		AvatarURL:   "https://example.com/avatar.png",
		DeviceInfo:  "test-device",
	}

	if metadata.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
	if metadata.DisplayName == "" {
		t.Error("display name should not be empty")
	}
}

func TestChallenge(t *testing.T) {
	challenge := &Challenge{
		Challenge: "test-challenge-string",
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if challenge.Challenge == "" {
		t.Error("challenge should not be empty")
	}
	if challenge.ExpiresAt.Before(challenge.Timestamp) {
		t.Error("expires at should be after timestamp")
	}
}

func TestChallengeResponse(t *testing.T) {
	response := &ChallengeResponse{
		Challenge: "test-challenge",
		Response:  "test-response",
		PeerID:    "test-peer-id",
	}

	if response.Challenge == "" {
		t.Error("challenge should not be empty")
	}
	if response.Response == "" {
		t.Error("response should not be empty")
	}
	if response.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrIdentityNotFound, "ErrIdentityNotFound"},
		{ErrIdentityExists, "ErrIdentityExists"},
		{ErrInvalidPeerID, "ErrInvalidPeerID"},
		{ErrInvalidPublicKey, "ErrInvalidPublicKey"},
		{ErrInvalidPrivateKey, "ErrInvalidPrivateKey"},
		{ErrKeyGenerationFailed, "ErrKeyGenerationFailed"},
		{ErrSignatureFailed, "ErrSignatureFailed"},
		{ErrVerificationFailed, "ErrVerificationFailed"},
		{ErrEncryptionFailed, "ErrEncryptionFailed"},
		{ErrDecryptionFailed, "ErrDecryptionFailed"},
		{ErrPasswordRequired, "ErrPasswordRequired"},
		{ErrPasswordMismatch, "ErrPasswordMismatch"},
		{ErrPasswordTooShort, "ErrPasswordTooShort"},
		{ErrInvalidFormat, "ErrInvalidFormat"},
		{ErrStorageFailed, "ErrStorageFailed"},
		{ErrChallengeExpired, "ErrChallengeExpired"},
		{ErrChallengeInvalid, "ErrChallengeInvalid"},
		{ErrMetadataNotFound, "ErrMetadataNotFound"},
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

func TestIdentityError(t *testing.T) {
	innerErr := ErrKeyGenerationFailed
	identityErr := NewIdentityError("KEY_GEN", "key generation failed", innerErr)

	if identityErr.Code != "KEY_GEN" {
		t.Errorf("expected code KEY_GEN, got %s", identityErr.Code)
	}
	if identityErr.Message != "key generation failed" {
		t.Errorf("expected message 'key generation failed', got %s", identityErr.Message)
	}
	if identityErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if identityErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ IdentityManager = nil
	var _ PeerIDUtil = nil
	var _ KeyStorage = nil
	var _ IdentityStore = nil
}
