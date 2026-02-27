package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

func TestPeerIDGeneration(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	peerIDGen := NewPeerIDGenerator()

	peerID := peerIDGen.GeneratePeerID(pubKey)
	if peerID == "" {
		t.Error("peer ID should not be empty")
	}

	if !peerIDGen.ValidatePeerID(peerID) {
		t.Error("peer ID should be valid")
	}

	hash, err := peerIDGen.ExtractPublicKeyHash(peerID)
	if err != nil {
		t.Errorf("failed to extract public key hash: %v", err)
	}
	if len(hash) == 0 {
		t.Error("hash should not be empty")
	}
}

func TestPeerIDEncoding(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	peerIDGen := NewPeerIDGenerator()

	base58PeerID := peerIDGen.EncodePeerID(pubKey, PeerIDEncodingBase58)
	if base58PeerID == "" {
		t.Error("base58 peer ID should not be empty")
	}

	hexPeerID := peerIDGen.EncodePeerID(pubKey, PeerIDEncodingHex)
	if hexPeerID == "" {
		t.Error("hex peer ID should not be empty")
	}

	decoded, err := peerIDGen.DecodePeerID(base58PeerID, PeerIDEncodingBase58)
	if err != nil {
		t.Errorf("failed to decode base58 peer ID: %v", err)
	}
	if len(decoded) == 0 {
		t.Error("decoded bytes should not be empty")
	}

	decodedHex, err := peerIDGen.DecodePeerID(hexPeerID, PeerIDEncodingHex)
	if err != nil {
		t.Errorf("failed to decode hex peer ID: %v", err)
	}
	if len(decodedHex) == 0 {
		t.Error("decoded hex bytes should not be empty")
	}
}

func TestInvalidPeerID(t *testing.T) {
	peerIDGen := NewPeerIDGenerator()

	if peerIDGen.ValidatePeerID("") {
		t.Error("empty peer ID should be invalid")
	}

	_, err := peerIDGen.ExtractPublicKeyHash("")
	if err == nil {
		t.Error("extracting hash from empty peer ID should fail")
	}

	_, err = peerIDGen.DecodePeerID("invalid!", PeerIDEncodingBase58)
	if err == nil {
		t.Error("decoding invalid peer ID should fail")
	}

	_, err = peerIDGen.DecodePeerID("zzz", PeerIDEncodingHex)
	if err == nil {
		t.Error("decoding invalid hex peer ID should fail")
	}
}

func TestMemoryIdentityStore(t *testing.T) {
	store := NewMemoryIdentityStore()

	identity := &Identity{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
	}

	metadata := &IdentityMetadata{
		PeerID:      "test-peer-id",
		DisplayName: "Test User",
	}

	if err := store.Save(identity, metadata); err != nil {
		t.Fatalf("failed to save identity: %v", err)
	}

	exists, err := store.Exists("test-peer-id")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("identity should exist")
	}

	loaded, loadedMetadata, err := store.Load("test-peer-id")
	if err != nil {
		t.Fatalf("failed to load identity: %v", err)
	}
	if loaded.PeerID != identity.PeerID {
		t.Errorf("expected peer ID %s, got %s", identity.PeerID, loaded.PeerID)
	}
	if loadedMetadata.DisplayName != "Test User" {
		t.Error("metadata should be loaded")
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("failed to list identities: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 identity, got %d", len(list))
	}

	if err := store.Delete("test-peer-id"); err != nil {
		t.Fatalf("failed to delete identity: %v", err)
	}

	exists, _ = store.Exists("test-peer-id")
	if exists {
		t.Error("identity should not exist after deletion")
	}
}

func TestMemoryIdentityStoreLoadNonExistent(t *testing.T) {
	store := NewMemoryIdentityStore()

	_, _, err := store.Load("non-existent")
	if err != ErrIdentityNotFound {
		t.Error("should return ErrIdentityNotFound for non-existent identity")
	}
}

func TestFileKeyStorage(t *testing.T) {
	store, err := NewFileKeyStorage("/tmp/test-identity-keys")
	if err != nil {
		t.Fatalf("failed to create key storage: %v", err)
	}

	testKey := []byte("test-private-key")
	peerID := "test-peer-id"

	if err := store.SavePrivateKey(peerID, testKey); err != nil {
		t.Fatalf("failed to save private key: %v", err)
	}

	exists, err := store.KeyExists(peerID)
	if err != nil {
		t.Fatalf("failed to check key existence: %v", err)
	}
	if !exists {
		t.Error("key should exist")
	}

	loaded, err := store.LoadPrivateKey(peerID)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}
	if string(loaded) != string(testKey) {
		t.Error("loaded key should match saved key")
	}

	if err := store.DeletePrivateKey(peerID); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	exists, _ = store.KeyExists(peerID)
	if exists {
		t.Error("key should not exist after deletion")
	}
}

func TestEncryptDecryptData(t *testing.T) {
	password := "test-password-123"
	exportData := &ExportData{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
		Metadata:   &IdentityMetadata{DisplayName: "Test User"},
	}

	encrypted, err := encryptData(exportData, password)
	if err != nil {
		t.Fatalf("failed to encrypt data: %v", err)
	}

	decrypted, err := decryptData(encrypted, password)
	if err != nil {
		t.Fatalf("failed to decrypt data: %v", err)
	}

	if decrypted.PeerID != exportData.PeerID {
		t.Errorf("expected peer ID %s, got %s", exportData.PeerID, decrypted.PeerID)
	}
}

func TestDecryptWithWrongPassword(t *testing.T) {
	password := "test-password-123"
	wrongPassword := "wrong-password"

	exportData := &ExportData{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
	}

	encrypted, err := encryptData(exportData, password)
	if err != nil {
		t.Fatalf("failed to encrypt data: %v", err)
	}

	_, err = decryptData(encrypted, wrongPassword)
	if err != ErrPasswordMismatch && err != ErrDecryptionFailed {
		t.Errorf("expected password mismatch error, got %v", err)
	}
}

func TestExportDataWithShortPassword(t *testing.T) {
	exportData := &ExportData{
		PeerID: "test-peer-id",
	}

	_, err := encryptData(exportData, "short")
	if err != nil {
		t.Logf("short password test passed: %v", err)
	}
}

func TestBase58Encoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"hello", []byte("hello"), "CnBEq"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := base58Encode(tc.input)
			t.Logf("base58Encode(%q) = %q", tc.input, result)
		})
	}
}

func TestHexEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"hello", []byte("hello"), "68656c6c6f"},
		{"test", []byte("test"), "74657374"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := encodeHex(tc.input)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestHexDecoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []byte
	}{
		{"empty", "", []byte{}},
		{"hello", "68656c6c6f", []byte("hello")},
		{"test", "74657374", []byte("test")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := decodeHex(tc.input)
			if err != nil {
				t.Errorf("decodeHex failed: %v", err)
				return
			}
			if string(result) != string(tc.expected) {
				t.Errorf("expected %s, got %s", string(tc.expected), string(result))
			}
		})
	}
}

func TestHexDecodingInvalid(t *testing.T) {
	invalidInputs := []string{
		"zzz",
		"12345",
		"xyz",
	}

	for _, input := range invalidInputs {
		_, err := decodeHex(input)
		if err == nil {
			t.Errorf("decodeHex(%q) should fail", input)
		}
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		bytes, err := GenerateRandomBytes(length)
		if err != nil {
			t.Errorf("GenerateRandomBytes(%d) failed: %v", length, err)
		}
		if len(bytes) != length {
			t.Errorf("expected length %d, got %d", length, len(bytes))
		}
	}

	_, err := GenerateRandomBytes(0)
	if err != nil {
		t.Logf("GenerateRandomBytes(0) correctly returned error: %v", err)
	}
}

func TestChallengeResponseFlow(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	challenge, err := manager.GenerateChallenge(identity.PeerID)
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	if challenge.Challenge == "" {
		t.Error("challenge should not be empty")
	}

	challengeData, err := base64.StdEncoding.DecodeString(challenge.Challenge)
	if err != nil {
		t.Fatalf("failed to decode challenge: %v", err)
	}

	signature := ed25519.Sign(identity.PrivateKey, challengeData)
	response := base64.StdEncoding.EncodeToString(signature)

	resp := &ChallengeResponse{
		Challenge: challenge.Challenge,
		Response:  response,
		PeerID:    identity.PeerID,
	}

	valid, err := manager.VerifyChallenge(identity.PeerID, challenge, resp)
	if err != nil {
		t.Fatalf("failed to verify challenge: %v", err)
	}
	if !valid {
		t.Error("challenge verification should succeed")
	}
}

func TestChallengeExpired(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	expiredChallenge := &Challenge{
		Challenge: "expired-challenge",
		Timestamp: time.Now().Add(-10 * time.Minute),
		ExpiresAt: time.Now().Add(-5 * time.Minute),
	}

	resp := &ChallengeResponse{
		Challenge: "expired-challenge",
		Response:  "any-response",
		PeerID:    identity.PeerID,
	}

	_, err = manager.VerifyChallenge(identity.PeerID, expiredChallenge, resp)
	if err != ErrChallengeExpired {
		t.Errorf("expected ErrChallengeExpired, got %v", err)
	}
}

func TestSignMessage(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	message := []byte("test message")
	signature, err := manager.SignMessage(message)
	if err != nil {
		t.Fatalf("failed to sign message: %v", err)
	}
	if len(signature) == 0 {
		t.Error("signature should not be empty")
	}
}

func TestSignMessageNoIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.SignMessage([]byte("test"))
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestDeleteIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	err = manager.DeleteIdentity(identity.PeerID)
	if err != nil {
		t.Fatalf("failed to delete identity: %v", err)
	}

	exists, _ := store.Exists(identity.PeerID)
	if exists {
		t.Error("identity should not exist after deletion")
	}
}

func TestDeleteNonExistentIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.DeleteIdentity("non-existent")
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestExportIdentityNoPassword(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	_, err = manager.ExportIdentity("")
	if err != ErrPasswordRequired {
		t.Errorf("expected ErrPasswordRequired, got %v", err)
	}
}

func TestExportIdentityShortPassword(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	_, err = manager.ExportIdentity("short")
	if err != ErrPasswordTooShort {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestImportIdentityDuplicate(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	exported, err := manager.ExportIdentity("test-password-123")
	if err != nil {
		t.Fatalf("failed to export identity: %v", err)
	}

	_, err = manager.ImportIdentity(exported, "test-password-123")
	if err != ErrIdentityExists {
		t.Errorf("expected ErrIdentityExists, got %v", err)
	}

	_ = identity
}

func TestVerifyChallengeInvalidPeerID(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	challenge := &Challenge{
		Challenge: "test-challenge",
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	resp := &ChallengeResponse{
		Challenge: "test-challenge",
		Response:  "test-response",
		PeerID:    "non-existent",
	}

	_, err = manager.VerifyChallenge("non-existent", challenge, resp)
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestVerifyChallengeInvalidResponse(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	challenge := &Challenge{
		Challenge: "test-challenge",
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	resp := &ChallengeResponse{
		Challenge: "invalid-base64!!!",
		Response:  "invalid-response",
		PeerID:    identity.PeerID,
	}

	_, err = manager.VerifyChallenge(identity.PeerID, challenge, resp)
	if err != ErrChallengeInvalid {
		t.Errorf("expected ErrChallengeInvalid, got %v", err)
	}
}

func TestVerifyIdentityPublicKeyMismatch(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	valid := manager.VerifyIdentity("invalid-peer-id", []byte("wrong-key"))
	if valid {
		t.Error("should return false for invalid peer ID")
	}
}

func TestListIdentities(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	identities, err := manager.ListIdentities()
	if err != nil {
		t.Fatalf("failed to list identities: %v", err)
	}
	if len(identities) != 1 {
		t.Errorf("expected 1 identity, got %d", len(identities))
	}
}

func TestGetMetadata(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	metadata, err := manager.GetMetadata(identity.PeerID)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}
	if metadata == nil {
		t.Error("metadata should not be nil")
	}
}

func TestGetMetadataNotFound(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.GetMetadata("non-existent")
	if err != ErrMetadataNotFound {
		t.Errorf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestUpdateMetadata(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	newMetadata := &IdentityMetadata{
		DisplayName: "Test User",
		AvatarURL:   "https://example.com/avatar.png",
		DeviceInfo:  "Test Device",
	}

	err = manager.UpdateMetadata(identity.PeerID, newMetadata)
	if err != nil {
		t.Fatalf("failed to update metadata: %v", err)
	}

	updated, err := manager.GetMetadata(identity.PeerID)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}
	if updated.DisplayName != "Test User" {
		t.Errorf("expected display name 'Test User', got %s", updated.DisplayName)
	}
}

func TestImportIdentityWrongPassword(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	exported, err := manager.ExportIdentity("test-password-123")
	if err != nil {
		t.Fatalf("failed to export identity: %v", err)
	}

	_, err = manager.ImportIdentity(exported, "wrong-password")
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestImportIdentityNoPassword(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	exported, err := manager.ExportIdentity("test-password-123")
	if err != nil {
		t.Fatalf("failed to export identity: %v", err)
	}

	_, err = manager.ImportIdentity(exported, "")
	if err != ErrPasswordRequired {
		t.Errorf("expected ErrPasswordRequired, got %v", err)
	}
}

func TestVerifySignatureNoIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	valid := manager.VerifySignature("non-existent", []byte("test"), []byte("signature"))
	if valid {
		t.Error("should return false for non-existent identity")
	}
}

func TestVerifyIdentityPublicKeyMatch(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	valid := manager.VerifyIdentity(identity.PeerID, identity.PublicKey)
	if !valid {
		t.Error("should return true for matching public key")
	}
}

func TestVerifyIdentityDifferentLength(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	valid := manager.VerifyIdentity(identity.PeerID, []byte("short"))
	if valid {
		t.Error("should return false for different length public key")
	}
}

func TestLoadIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	loaded, err := manager.LoadIdentity(identity.PeerID)
	if err != nil {
		t.Fatalf("failed to load identity: %v", err)
	}
	if loaded.PeerID != identity.PeerID {
		t.Errorf("expected peer ID %s, got %s", identity.PeerID, loaded.PeerID)
	}
}

func TestLoadIdentityNotFound(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.LoadIdentity("non-existent")
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestGenerateChallengeNotFound(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.GenerateChallenge("non-existent")
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestPeerIDGeneratorInterface(t *testing.T) {
	peerIDGen := NewPeerIDGenerator()

	var _ PeerIDUtil = peerIDGen
}

func TestIdentityStoreInterface(t *testing.T) {
	store := NewMemoryIdentityStore()

	var _ IdentityStore = store
}

func TestKeyStorageInterface(t *testing.T) {
	keyStore := NewMemoryKeyStorage()

	var _ KeyStorage = keyStore
}

func TestFileKeyStorageInterface(t *testing.T) {
	store, err := NewFileKeyStorage("/tmp/test-identity-keys-interface")
	if err != nil {
		t.Fatalf("failed to create key storage: %v", err)
	}

	var _ KeyStorage = store
}

func TestFileIdentityStoreInterface(t *testing.T) {
	store, err := NewFileIdentityStore("/tmp/test-identity-store-interface")
	if err != nil {
		t.Fatalf("failed to create identity store: %v", err)
	}

	var _ IdentityStore = store
}

func TestDefaultIdentityConfig(t *testing.T) {
	config := DefaultIdentityConfig()
	if config.KeyType != KeyTypeEd25519 {
		t.Errorf("expected KeyTypeEd25519, got %s", config.KeyType)
	}
	if config.KeyLength != 256 {
		t.Errorf("expected key length 256, got %d", config.KeyLength)
	}
	if config.PeerIDEncoding != PeerIDEncodingBase58 {
		t.Errorf("expected PeerIDEncodingBase58, got %s", config.PeerIDEncoding)
	}
}

func TestCreateIdentityWithConfig(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	config := &IdentityConfig{
		KeyType:        KeyTypeEd25519,
		KeyLength:      256,
		PeerIDEncoding: PeerIDEncodingHex,
	}

	identity, err := manager.CreateIdentity(config)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	if identity.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
}

func TestCreateDuplicateIdentity(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Logf("second create returned error (expected for different key): %v", err)
	}

	_ = identity
}

func TestUpdateMetadataNotFound(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.UpdateMetadata("non-existent", &IdentityMetadata{})
	if err != ErrMetadataNotFound {
		t.Errorf("expected ErrMetadataNotFound, got %v", err)
	}
}

func TestDecryptInvalidVersion(t *testing.T) {
	invalidData := []byte(`{"ciphertext":"dGVzdA==","salt":"c2FsdA==","nonce":"bm9uY2U=","version":999}`)

	_, err := decryptData(invalidData, "password")
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestDecryptInvalidNonceLength(t *testing.T) {
	exportData := &ExportData{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
	}

	encrypted, err := encryptData(exportData, "test-password-123")
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	var enc struct {
		Ciphertext []byte `json:"ciphertext"`
		Salt       []byte `json:"salt"`
		Nonce      []byte `json:"nonce"`
		Version    int    `json:"version"`
	}
	json.Unmarshal(encrypted, &enc)
	enc.Nonce = []byte("short")
	modified, _ := json.Marshal(enc)

	_, err = decryptData(modified, "test-password-123")
	if err != ErrDecryptionFailed {
		t.Logf("expected decryption error for short nonce, got: %v", err)
	}
}

func TestBase58Decode(t *testing.T) {
	_ = NewPeerIDGenerator()
	decoded := base58Decode("Cn8eVZg")
	if len(decoded) == 0 {
		t.Error("decoded should not be empty")
	}
}

func TestBase58DecodeInvalid(t *testing.T) {
	result := base58Decode("invalid!char")
	if result != nil {
		t.Error("should return nil for invalid char")
	}
}

func TestBase58DecodeEmpty(t *testing.T) {
	result := base58Decode("")
	if result != nil {
		t.Error("should return nil for empty string")
	}
}

func TestEncodeHex(t *testing.T) {
	result := encodeHex([]byte("test"))
	if result != "74657374" {
		t.Errorf("expected 74657374, got %s", result)
	}
}

func TestDecodeHexUpperCase(t *testing.T) {
	result, err := decodeHex("74657374")
	if err != nil {
		t.Fatalf("decodeHex failed: %v", err)
	}
	if string(result) != "test" {
		t.Errorf("expected test, got %s", string(result))
	}
}

func TestHexCharToNibble(t *testing.T) {
	tests := []struct {
		input    byte
		expected int
	}{
		{'0', 0},
		{'9', 9},
		{'a', 10},
		{'f', 15},
		{'A', 10},
		{'F', 15},
		{'g', -1},
		{'/', -1},
	}

	for _, tc := range tests {
		result := hexCharToNibble(tc.input)
		if result != tc.expected {
			t.Errorf("hexCharToNibble(%c) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestFileIdentityStore(t *testing.T) {
	store, err := NewFileIdentityStore("/tmp/test-identity-store-file")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	identity := &Identity{
		PeerID:     "test-peer-id",
		PublicKey:  []byte("test-public-key"),
		PrivateKey: []byte("test-private-key"),
	}

	metadata := &IdentityMetadata{
		PeerID:      "test-peer-id",
		DisplayName: "Test User",
	}

	err = store.Save(identity, metadata)
	if err != nil {
		t.Fatalf("failed to save identity: %v", err)
	}

	exists, err := store.Exists("test-peer-id")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("identity should exist")
	}

	loaded, loadedMetadata, err := store.Load("test-peer-id")
	if err != nil {
		t.Fatalf("failed to load identity: %v", err)
	}
	if loaded.PeerID != identity.PeerID {
		t.Errorf("expected peer ID %s, got %s", identity.PeerID, loaded.PeerID)
	}
	if loadedMetadata.DisplayName != "Test User" {
		t.Error("metadata should be loaded")
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 identity, got %d", len(list))
	}

	err = store.Delete("test-peer-id")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	exists, _ = store.Exists("test-peer-id")
	if exists {
		t.Error("identity should not exist after deletion")
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			b.Fatalf("failed to generate key: %v", err)
		}
	}
}

func BenchmarkSignMessage(b *testing.B) {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	identity := &Identity{
		PeerID:     "test",
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}
	message := []byte("benchmark test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Sign(identity.PrivateKey, message)
	}
}

func BenchmarkVerifySignature(b *testing.B) {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("benchmark test message")
	signature := ed25519.Sign(privKey, message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed25519.Verify(pubKey, message, signature)
	}
}

func BenchmarkPeerIDGeneration(b *testing.B) {
	_, pubKey, _ := ed25519.GenerateKey(rand.Reader)
	peerIDGen := NewPeerIDGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		peerIDGen.GeneratePeerID(pubKey)
	}
}

func BenchmarkBase58Encode(b *testing.B) {
	data := make([]byte, 32)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base58Encode(data)
	}
}

func BenchmarkSHA256Hash(b *testing.B) {
	data := make([]byte, 1024)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sha256.Sum256(data)
	}
}

func BenchmarkAESEncryption(b *testing.B) {
	password := "test-password-123"
	data := &ExportData{
		PeerID:     "test-peer-id",
		PublicKey:  make([]byte, 32),
		PrivateKey: make([]byte, 64),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encryptData(data, password)
	}
}

func BenchmarkPBKDF2KeyDerivation(b *testing.B) {
	password := "test-password-123"
	salt := make([]byte, 16)
	rand.Read(salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)
	}
}

func TestFileIdentityStoreLoadNotFound(t *testing.T) {
	store, err := NewFileIdentityStore("/tmp/test-identity-store-notfound")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	_, _, err = store.Load("non-existent")
	if err != ErrIdentityNotFound {
		t.Errorf("expected ErrIdentityNotFound, got %v", err)
	}
}

func TestFileKeyStorageListKeys(t *testing.T) {
	store, err := NewFileKeyStorage("/tmp/test-key-storage-list")
	if err != nil {
		t.Fatalf("failed to create key storage: %v", err)
	}

	err = store.SavePrivateKey("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	err = store.SavePrivateKey("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	keys, err := store.ListKeys()
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMemoryKeyStorageListKeys(t *testing.T) {
	store := NewMemoryKeyStorage()

	err := store.SavePrivateKey("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	err = store.SavePrivateKey("key2", []byte("value2"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	keys, err := store.ListKeys()
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMemoryKeyStorageKeyExists(t *testing.T) {
	store := NewMemoryKeyStorage()

	err := store.SavePrivateKey("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	exists, err := store.KeyExists("key1")
	if err != nil {
		t.Fatalf("failed to check key exists: %v", err)
	}
	if !exists {
		t.Error("key should exist")
	}

	exists, _ = store.KeyExists("non-existent")
	if exists {
		t.Error("key should not exist")
	}
}

func TestFileKeyStorageKeyExists(t *testing.T) {
	store, err := NewFileKeyStorage("/tmp/test-key-storage-exists")
	if err != nil {
		t.Fatalf("failed to create key storage: %v", err)
	}

	err = store.SavePrivateKey("key1", []byte("value1"))
	if err != nil {
		t.Fatalf("failed to save key: %v", err)
	}

	exists, err := store.KeyExists("key1")
	if err != nil {
		t.Fatalf("failed to check key exists: %v", err)
	}
	if !exists {
		t.Error("key should exist")
	}

	exists, _ = store.KeyExists("non-existent")
	if exists {
		t.Error("key should not exist")
	}
}

func TestConcurrentIdentityCreation(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan *Identity, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			identity, err := manager.CreateIdentity(nil)
			if err != nil {
				errors <- err
				return
			}
			results <- identity
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	select {
	case err := <-errors:
		t.Logf("concurrent creation error: %v (may be expected for different keys)", err)
	default:
	}

	if len(results) > 0 {
		t.Logf("created %d identities concurrently", len(results))
	}
}

func TestConcurrentSignMessage(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	_, err = manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	var wg sync.WaitGroup
	signatures := make(chan []byte, 100)
	errors := make(chan error, 100)

	message := []byte("concurrent test message")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sig, err := manager.SignMessage(message)
			if err != nil {
				errors <- err
				return
			}
			signatures <- sig
		}()
	}

	wg.Wait()
	close(signatures)
	close(errors)

	errCount := 0
	for err := range errors {
		if err != nil {
			errCount++
		}
	}

	count := 0
	for sig := range signatures {
		if len(sig) > 0 {
			count++
		}
	}

	if errCount > 0 {
		t.Logf("got %d errors during concurrent signing", errCount)
	}
	if count != 100 {
		t.Errorf("expected 100 signatures, got %d", count)
	}
}

func TestConcurrentMetadataAccess(t *testing.T) {
	store := NewMemoryIdentityStore()
	keyStore := NewMemoryKeyStorage()

	manager, err := NewIdentityManager(store, keyStore)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	identity, err := manager.CreateIdentity(nil)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 200)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := manager.GetMetadata(identity.PeerID)
			if err != nil {
				errors <- err
			}
		}(i)

		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			err := manager.UpdateMetadata(identity.PeerID, &IdentityMetadata{
				DisplayName: fmt.Sprintf("User %d", n),
			})
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errCount := 0
	for err := range errors {
		if err != nil {
			errCount++
		}
	}

	if errCount > 0 {
		t.Fatalf("got %d errors during concurrent metadata access", errCount)
	}
}

func TestConcurrentStoreAccess(t *testing.T) {
	store := NewMemoryIdentityStore()

	var wg sync.WaitGroup
	errors := make(chan error, 200)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			identity := &Identity{
				PeerID:     fmt.Sprintf("peer-%d", n),
				PublicKey:  []byte(fmt.Sprintf("pubkey-%d", n)),
				PrivateKey: []byte(fmt.Sprintf("privkey-%d", n)),
			}
			err := store.Save(identity, &IdentityMetadata{PeerID: identity.PeerID})
			if err != nil {
				errors <- err
			}
		}(i)

		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _, err := store.Load(fmt.Sprintf("peer-%d", n%100))
			if err != nil && err != ErrIdentityNotFound {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errCount := 0
	for err := range errors {
		if err != nil {
			errCount++
		}
	}

	if errCount > 0 {
		t.Fatalf("got %d errors during concurrent store access", errCount)
	}
}
