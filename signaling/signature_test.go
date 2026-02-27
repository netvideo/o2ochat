package signaling

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"
)

func TestNewMessageSigner(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	peerID := "test-peer"
	signer := NewMessageSigner(privateKey, peerID)

	if signer == nil {
		t.Fatal("Expected signer to be created")
	}

	if signer.GetPeerID() != peerID {
		t.Errorf("Expected peer ID %s, got %s", peerID, signer.GetPeerID())
	}

	if len(signer.GetPublicKey()) != ed25519.PublicKeySize {
		t.Errorf("Expected public key length %d, got %d", ed25519.PublicKeySize, len(signer.GetPublicKey()))
	}
}

func TestMessageSignerSignAndVerify(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	signer := NewMessageSigner(privateKey, "test-peer")

	message := []byte("test message")
	signature, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	if len(signature) != ed25519.SignatureSize {
		t.Errorf("Expected signature length %d, got %d", ed25519.SignatureSize, len(signature))
	}

	hash := sha256.Sum256(message)
	valid := ed25519.Verify(publicKey, hash[:], signature)
	if !valid {
		t.Error("Signature verification failed")
	}
}

func TestMessageSignerInvalidSignature(t *testing.T) {
	_, privateKey1, _ := ed25519.GenerateKey(rand.Reader)
	_, privateKey2, _ := ed25519.GenerateKey(rand.Reader)

	signer1 := NewMessageSigner(privateKey1, "peer1")
	signer2 := NewMessageSigner(privateKey2, "peer2")

	message := []byte("test message")
	signature, _ := signer1.SignMessage(message)

	valid := ed25519.Verify(signer2.GetPublicKey(), message, signature)
	if valid {
		t.Error("Expected signature verification to fail with different key")
	}
}

func TestNonceStore(t *testing.T) {
	store := NewNonceStore()

	nonce := "test-nonce"
	result := store.Add(nonce, 1*time.Second)
	if !result {
		t.Error("Expected nonce to be added")
	}

	valid := store.Validate(nonce)
	if !valid {
		t.Error("Expected nonce to be valid")
	}
}

func TestNonceStoreDuplicate(t *testing.T) {
	store := NewNonceStore()

	nonce := "test-nonce"
	store.Add(nonce, 1*time.Second)

	result := store.Add(nonce, 1*time.Second)
	if result {
		t.Error("Expected duplicate nonce to be rejected")
	}
}

func TestNonceStoreExpired(t *testing.T) {
	store := NewNonceStore()

	nonce := "test-nonce"
	store.Add(nonce, 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	valid := store.Validate(nonce)
	if valid {
		t.Error("Expected expired nonce to be invalid")
	}
}

func TestNonceStoreCleanup(t *testing.T) {
	store := NewNonceStore()

	store.Add("nonce1", 10*time.Millisecond)
	store.Add("nonce2", 1*time.Hour)

	time.Sleep(20 * time.Millisecond)
	store.Cleanup()

	if _, exists := store.nonces["nonce1"]; exists {
		t.Error("Expected expired nonce to be cleaned up")
	}

	if _, exists := store.nonces["nonce2"]; !exists {
		t.Error("Expected valid nonce to remain")
	}
}

func TestSignMessage(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	msg := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		To:        "peer2",
		Timestamp: time.Now(),
		Nonce:     "test-nonce",
	}

	signature, err := SignMessage(msg, privateKey)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	msg.Signature = signature
	valid := VerifySignature(msg, publicKey)
	if !valid {
		t.Error("Signature verification failed")
	}
}

func TestVerifySignatureNoSignature(t *testing.T) {
	publicKey, _, _ := ed25519.GenerateKey(rand.Reader)

	msg := &SignalingMessage{
		Type: MessageTypePing,
	}

	valid := VerifySignature(msg, publicKey)
	if valid {
		t.Error("Expected verification to fail with no signature")
	}
	}

	invalidNonce := "invalid"
	if ValidateNonceFormat(invalidNonce) {
		t.Error("Expected invalid nonce format to be rejected")
	}

	emptyNonce := ""
	if ValidateNonceFormat(emptyNonce) {
		t.Error("Expected empty nonce to be rejected")
	}
}

func TestGetSignableData(t *testing.T) {
	msg := &SignalingMessage{
		Type:      MessageTypeOffer,
		From:      "peer1",
		To:        "peer2",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
		Nonce:     "test-nonce",
	}

	data, err := GetSignableData(msg)
	if err != nil {
		t.Fatalf("GetSignableData failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty signable data")
	}
}

func TestCreateSignableMessage(t *testing.T) {
	msg := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		Timestamp: time.Now(),
	}

	data, err := CreateSignableMessage(msg)
	if err != nil {
		t.Fatalf("CreateSignableMessage failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty signable message")
	}
}

func TestHashMessage(t *testing.T) {
	msg1 := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		Timestamp: time.Now(),
	}

	hash1, err := HashMessage(msg1)
	if err != nil {
		t.Fatalf("HashMessage failed: %v", err)
	}

	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	msg2 := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		Timestamp: msg1.Timestamp,
	}

	hash2, _ := HashMessage(msg2)
	if hash1 != hash2 {
		t.Error("Expected same hash for identical messages")
	}
}

func TestValidateMessageTimestamp(t *testing.T) {
	now := time.Now()

	msg := &SignalingMessage{
		Timestamp: now,
	}

	valid := ValidateMessageTimestamp(msg, 1*time.Minute)
	if !valid {
		t.Error("Expected current timestamp to be valid")
	}

	oldMsg := &SignalingMessage{
		Timestamp: now.Add(-2 * time.Minute),
	}

	valid = ValidateMessageTimestamp(oldMsg, 1*time.Minute)
	if valid {
		t.Error("Expected old timestamp to be invalid")
	}

	futureMsg := &SignalingMessage{
		Timestamp: now.Add(2 * time.Minute),
	}

	valid = ValidateMessageTimestamp(futureMsg, 1*time.Minute)
	if valid {
		t.Error("Expected future timestamp to be invalid")
	}
}

func TestSignMessageToBase64(t *testing.T) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	msg := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		Timestamp: time.Now(),
	}

	signatureBase64, err := SignMessageToBase64(msg, privateKey)
	if err != nil {
		t.Fatalf("SignMessageToBase64 failed: %v", err)
	}

	if signatureBase64 == "" {
		t.Error("Expected non-empty base64 signature")
	}
}

func TestVerifySignatureFromBase64(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	msg := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		Timestamp: time.Now(),
	}

	signatureBase64, _ := SignMessageToBase64(msg, privateKey)

	valid := VerifySignatureFromBase64(msg, publicKey, signatureBase64)
	if !valid {
		t.Error("Expected signature verification to succeed")
	}
}

func TestVerifySignatureFromBase64Invalid(t *testing.T) {
	publicKey, _, _ := ed25519.GenerateKey(rand.Reader)

	msg := &SignalingMessage{
		Type: MessageTypePing,
	}

	valid := VerifySignatureFromBase64(msg, publicKey, "invalid-base64")
	if valid {
		t.Error("Expected invalid base64 to fail verification")
	}
}

func TestMessageSignerAddNonce(t *testing.T) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	signer := NewMessageSigner(privateKey, "test-peer")

	result := signer.AddNonce("test-nonce", 1*time.Second)
	if !result {
		t.Error("Expected nonce to be added")
	}
}

func TestMessageSignerValidateNonce(t *testing.T) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	signer := NewMessageSigner(privateKey, "test-peer")

	signer.AddNonce("test-nonce", 1*time.Second)

	valid := signer.ValidateNonce("test-nonce")
	if !valid {
		t.Error("Expected nonce to be valid")
	}

	invalid := signer.ValidateNonce("non-existent-nonce")
	if invalid {
		t.Error("Expected non-existent nonce to be invalid")
	}
}
