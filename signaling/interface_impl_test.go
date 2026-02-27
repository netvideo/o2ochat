//go:build !integration
// +build !integration

package signaling

import (
	"testing"
	"time"
)

func TestSignalingClientInterface(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)
	
	if client == nil {
		t.Fatal("Failed to create client")
	}
	
	var _ SignalingClient = client
	t.Log("Client implements SignalingClient interface")
	
	client.Close()
}

func TestSignalingServerInterface(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)
	
	if server == nil {
		t.Fatal("Failed to create server")
	}
	
	var _ SignalingServer = server
	t.Log("Server implements SignalingServer interface")
}

func TestMessageSignerInterface(t *testing.T) {
	signer := &MessageSigner{}
	
	var _ Signer = signer
	t.Log("MessageSigner implements Signer interface")
}

func TestNonceStoreBasic(t *testing.T) {
	store := NewNonceStore()
	
	if store == nil {
		t.Fatal("Failed to create nonce store")
	}
	
	nonce := "test-nonce-123"
	added := store.Add(nonce, 5*time.Second)
	if !added {
		t.Error("Failed to add nonce")
	}
	
	valid := store.Validate(nonce)
	if !valid {
		t.Error("Nonce should be valid")
	}
}

func TestNonceStoreDuplicateReject(t *testing.T) {
	store := NewNonceStore()
	nonce := "test-nonce-dup"
	
	store.Add(nonce, 5*time.Second)
	result := store.Add(nonce, 5*time.Second)
	
	if result {
		t.Error("Should reject duplicate nonce")
	}
}

func TestNonceStoreExpiry(t *testing.T) {
	store := NewNonceStore()
	nonce := "test-nonce-expiry"
	
	store.Add(nonce, 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	
	valid := store.Validate(nonce)
	if valid {
		t.Error("Expired nonce should be invalid")
	}
}

func TestDefaultServerConfigValues(t *testing.T) {
	config := DefaultServerConfig()
	
	if config.Port <= 0 {
		t.Error("Port should be positive")
	}
	if config.HeartbeatInterval <= 0 {
		t.Error("HeartbeatInterval should be positive")
	}
	if config.MessageTimeout <= 0 {
		t.Error("MessageTimeout should be positive")
	}
}

func TestDefaultClientConfigValues(t *testing.T) {
	config := DefaultClientConfig()
	
	if config.ServerURL == "" {
		t.Error("ServerURL should not be empty")
	}
	if config.HeartbeatInterval <= 0 {
		t.Error("HeartbeatInterval should be positive")
	}
	if config.MessageTimeout <= 0 {
		t.Error("MessageTimeout should be positive")
	}
}

func TestGenerateNonceUniqueness(t *testing.T) {
	nonces := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		nonce := generateNonce()
		if nonces[nonce] {
			t.Errorf("Duplicate nonce generated: %s", nonce)
		}
		nonces[nonce] = true
	}
}

func TestMessageTypeValues(t *testing.T) {
	types := map[MessageType]string{
		MessageTypeOffer:     "offer",
		MessageTypeAnswer:    "answer",
		MessageTypeCandidate: "candidate",
		MessageTypeInvite:    "invite",
		MessageTypeBye:       "bye",
		MessageTypePing:      "ping",
		MessageTypePong:      "pong",
	}
	
	for msgType, expected := range types {
		if string(msgType) != expected {
			t.Errorf("MessageType %s has wrong value", expected)
		}
	}
}

func TestConnectionStateValues(t *testing.T) {
	states := map[ConnectionState]string{
		StateDisconnected: "disconnected",
		StateConnecting:    "connecting",
		StateConnected:     "connected",
		StateReconnecting:  "reconnecting",
		StateFailed:        "failed",
	}
	
	for state, expected := range states {
		if state.String() != expected {
			t.Errorf("ConnectionState %s has wrong string value", expected)
		}
	}
}
