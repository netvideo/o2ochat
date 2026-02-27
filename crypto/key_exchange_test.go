package crypto

import (
	"testing"
)

func TestNewKeyExchange(t *testing.T) {
	ke := NewKeyExchange(nil)
	if ke == nil {
		t.Fatal("Expected key exchange to be created")
	}
}

func TestKeyExchangeInitiate(t *testing.T) {
	ke := NewKeyExchange(nil)

	result, err := ke.Initiate()
	if err != nil {
		t.Fatalf("Initiate failed: %v", err)
	}

	if len(result.PublicKey) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(result.PublicKey))
	}

	if result.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}

	if result.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}
}

func TestKeyExchangeRespond(t *testing.T) {
	ke := NewKeyExchange(nil)

	peerPublicKey := make([]byte, 32)
	for i := range peerPublicKey {
		peerPublicKey[i] = byte(i)
	}

	result, err := ke.Respond(peerPublicKey)
	if err != nil {
		t.Fatalf("Respond failed: %v", err)
	}

	if len(result.PublicKey) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(result.PublicKey))
	}

	if result.SharedSecret == nil {
		t.Error("Expected shared secret to be generated")
	}

	if len(result.SharedSecret) != 32 {
		t.Errorf("Expected shared secret length 32, got %d", len(result.SharedSecret))
	}
}

func TestKeyExchangeInvalidPublicKey(t *testing.T) {
	ke := NewKeyExchange(nil)

	_, err := ke.Respond([]byte("invalid"))
	if err != ErrInvalidPublicKey {
		t.Errorf("Expected ErrInvalidPublicKey, got: %v", err)
	}
}

func TestKeyExchangeFinalize(t *testing.T) {
	ke := NewKeyExchange(nil)

	initResult, _ := ke.Initiate()

	peerPublicKey := make([]byte, 32)
	for i := range peerPublicKey {
		peerPublicKey[i] = byte(i)
	}

	sharedSecret, err := ke.Finalize(peerPublicKey, initResult.SessionID)
	if err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}

	if sharedSecret == nil {
		t.Error("Expected shared secret")
	}
}

func TestKeyExchangeFinalizeNotFound(t *testing.T) {
	ke := NewKeyExchange(nil)

	_, err := ke.Finalize(make([]byte, 32), "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}
}

func TestKeyExchangeVerify(t *testing.T) {
	ke := NewKeyExchange(nil)

	sharedSecret := []byte("shared secret")
	proof := []byte("proof")

	_, err := ke.VerifyExchange(sharedSecret, proof)
	if err != nil {
		t.Logf("Verify returned: %v", err)
	}
}

func TestKeyExchangeDestroySession(t *testing.T) {
	ke := NewKeyExchange(nil)

	result, _ := ke.Initiate()

	err := ke.DestroySession(result.SessionID)
	if err != nil {
		t.Errorf("DestroySession failed: %v", err)
	}

	_, err = ke.Finalize(make([]byte, 32), result.SessionID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected session to be destroyed")
	}
}

func TestKeyExchangeDestroyNotFound(t *testing.T) {
	ke := NewKeyExchange(nil)

	err := ke.DestroySession("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}
}

func TestKeyExchangeMultiple(t *testing.T) {
	ke := NewKeyExchange(nil)

	results := make([]*KeyExchangeResult, 5)
	for i := 0; i < 5; i++ {
		result, err := ke.Initiate()
		if err != nil {
			t.Fatalf("Initiate %d failed: %v", i, err)
		}
		results[i] = result
	}

	sessionIDs := make(map[string]bool)
	for _, result := range results {
		if sessionIDs[result.SessionID] {
			t.Error("Expected unique session IDs")
		}
		sessionIDs[result.SessionID] = true
	}
}

func TestGenerateSessionID(t *testing.T) {
	data1 := []byte("test data 1")
	data2 := []byte("test data 2")

	id1 := generateSessionID(data1)
	id2 := generateSessionID(data2)

	if id1 == id2 {
		t.Error("Expected different session IDs")
	}

	id1Again := generateSessionID(data1)
	if id1 != id1Again {
		t.Error("Expected same session ID for same data")
	}
}
