package signaling

import (
	"testing"
	"time"
)

func TestNewWebSocketClient(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.GetState() != StateDisconnected {
		t.Error("Expected initial state to be disconnected")
	}
}

func TestNewWebSocketClientNilConfig(t *testing.T) {
	client := NewWebSocketClient(nil)

	if client == nil {
		t.Fatal("Expected client to be created with nil config")
	}

	if client.GetConfig() == nil {
		t.Error("Expected default config to be set")
	}
}

func TestClientConnectInvalidURL(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	err := client.Connect("invalid://url")
	if err == nil {
		t.Error("Expected error when connecting to invalid URL")
	}

	if client.GetState() != StateFailed {
		t.Errorf("Expected state to be failed, got %v", client.GetState())
	}

	client.Close()
}

func TestClientDisconnectNotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	err := client.Disconnect()
	if err != nil {
		t.Errorf("Expected nil, got: %v", err)
	}
}

func TestClientSendMessageNotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	msg := &SignalingMessage{
		Type: MessageTypePing,
	}

	err := client.SendMessage(msg)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}

	client.Close()
}

func TestClientReceiveMessageTimeout(t *testing.T) {
	config := &ClientConfig{
		ServerURL:            "ws://localhost:8080",
		HeartbeatInterval:    30 * time.Second,
		MessageTimeout:       100 * time.Millisecond,
		MaxReconnectAttempts: 0,
	}

	client := NewWebSocketClient(config)

	msg, err := client.ReceiveMessage()
	if err != ErrMessageTimeout && err != ErrNotConnected {
		t.Errorf("Expected timeout or not connected error, got: %v", err)
	}
	if msg != nil {
		t.Error("Expected nil message on timeout")
	}

	client.Close()
}

func TestClientGetState(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	state := client.GetState()
	if state != StateDisconnected {
		t.Errorf("Expected StateDisconnected, got %v", state)
	}

	client.Close()
}

func TestClientSetMessageHandler(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	called := false
	handler := func(msg *SignalingMessage) {
		called = true
	}

	client.SetMessageHandler(handler)

	if !called {
		t.Log("Handler set successfully (will be called when message received)")
	}

	client.Close()
}

func TestClientSetErrorHandler(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	called := false
	handler := func(err error) {
		called = true
		t.Logf("Error handler called with: %v", err)
	}

	client.SetErrorHandler(handler)

	if !called {
		t.Log("Handler set successfully (will be called when error occurs)")
	}

	client.Close()
}

func TestClientClose(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if client.GetState() != StateDisconnected {
		t.Error("Expected state to be disconnected after close")
	}
}

func TestClientMultipleClose(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	client.Close()
	client.Close()

	t.Log("Multiple close calls handled successfully")
}

func TestClientLookupPeerNotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	peer, err := client.LookupPeer("test-peer")
	if err == nil {
		t.Error("Expected error when looking up peer while not connected")
	}
	if peer != nil {
		t.Error("Expected nil peer on error")
	}

	client.Close()
}

func TestClientRegisterNotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	peerInfo := &PeerInfo{
		PeerID: "test-peer",
	}

	err := client.Register(peerInfo)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}

	client.Close()
}

func TestClientUnregisterNotConnected(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	err := client.Unregister()
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}

	client.Close()
}

func TestClientConfig(t *testing.T) {
	config := &ClientConfig{
		ServerURL:            "ws://localhost:8080",
		HeartbeatInterval:    30 * time.Second,
		MessageTimeout:       10 * time.Second,
		MaxReconnectAttempts: 3,
		ReconnectInterval:    5 * time.Second,
	}

	client := NewWebSocketClient(config)

	retrievedConfig := client.GetConfig()
	if retrievedConfig.ServerURL != config.ServerURL {
		t.Errorf("Expected server URL %s, got %s", config.ServerURL, retrievedConfig.ServerURL)
	}

	client.Close()
}

func TestClientDefaultConfig(t *testing.T) {
	config := DefaultClientConfig()

	if config.ServerURL == "" {
		t.Error("Expected default server URL to be set")
	}

	if config.HeartbeatInterval == 0 {
		t.Error("Expected default heartbeat interval to be set")
	}

	if config.MessageTimeout == 0 {
		t.Error("Expected default message timeout to be set")
	}
}

func TestClientStateTransitions(t *testing.T) {
	config := DefaultClientConfig()
	client := NewWebSocketClient(config)

	if client.GetState() != StateDisconnected {
		t.Error("Expected initial state to be disconnected")
	}

	client.Close()

	if client.GetState() != StateDisconnected {
		t.Error("Expected state to remain disconnected after close")
	}
}

func TestClientGenerateNonce(t *testing.T) {
	nonce1 := generateNonce()
	nonce2 := generateNonce()

	if nonce1 == "" {
		t.Error("Expected non-empty nonce")
	}

	if nonce1 == nonce2 {
		t.Error("Expected unique nonces")
	}
}
