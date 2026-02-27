package signaling

import (
	"testing"
	"time"
)

func TestNewWebSocketServer(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.GetConfig() != config {
		t.Error("Expected config to match")
	}

	if server.IsRunning() {
		t.Error("Expected server to not be running initially")
	}
}

func TestNewWebSocketServerNilConfig(t *testing.T) {
	server := NewWebSocketServer(nil)

	if server == nil {
		t.Fatal("Expected server to be created with nil config")
	}

	if server.GetConfig() == nil {
		t.Error("Expected default config to be set")
	}
}

func TestServerStartStop(t *testing.T) {
	config := &ServerConfig{
		Port:              0,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
	}

	server := NewWebSocketServer(config)

	err := server.Start(":0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running after Start")
	}

	peers, err := server.GetOnlinePeers()
	if err != nil {
		t.Errorf("GetOnlinePeers failed: %v", err)
	}
	if len(peers) != 0 {
		t.Errorf("Expected 0 peers, got %d", len(peers))
	}

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsRunning() {
		t.Error("Expected server to not be running after Stop")
	}
}

func TestServerStartAlreadyRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	err := server.Start(":0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	err = server.Start(":0")
	if err != ErrServerAlreadyRunning {
		t.Errorf("Expected ErrServerAlreadyRunning, got: %v", err)
	}

	server.Stop()
}

func TestServerStopNotRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	err := server.Stop()
	if err != ErrServerNotRunning {
		t.Errorf("Expected ErrServerNotRunning, got: %v", err)
	}
}

func TestServerBroadcast(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	err := server.Start(":0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	msg := &SignalingMessage{
		Type:      MessageTypePing,
		From:      "peer1",
		To:        "peer2",
		Timestamp: time.Now(),
	}

	err = server.Broadcast(msg)
	if err != nil {
		t.Errorf("Broadcast failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestServerBroadcastNotRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	msg := &SignalingMessage{
		Type: MessageTypePing,
	}

	err := server.Broadcast(msg)
	if err != ErrServerNotRunning {
		t.Errorf("Expected ErrServerNotRunning, got: %v", err)
	}
}

func TestServerGetOnlinePeers(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	err := server.Start(":0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	peers, err := server.GetOnlinePeers()
	if err != nil {
		t.Errorf("GetOnlinePeers failed: %v", err)
	}

	if len(peers) != 0 {
		t.Errorf("Expected 0 peers initially, got %d", len(peers))
	}
}

func TestServerGetConfig(t *testing.T) {
	config := &ServerConfig{
		Port:              8080,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
		MaxConnections:    1000,
	}

	server := NewWebSocketServer(config)

	retrievedConfig := server.GetConfig()
	if retrievedConfig.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, retrievedConfig.Port)
	}
}

func TestServerIsRunning(t *testing.T) {
	config := DefaultServerConfig()
	server := NewWebSocketServer(config)

	if server.IsRunning() {
		t.Error("Expected server to not be running initially")
	}

	err := server.Start(":0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running after Start")
	}

	server.Stop()

	if server.IsRunning() {
		t.Error("Expected server to not be running after Stop")
	}
}

func TestServerDefaultConfig(t *testing.T) {
	config := DefaultServerConfig()

	if config.Port == 0 {
		t.Error("Expected default port to be set")
	}

	if config.HeartbeatInterval == 0 {
		t.Error("Expected default heartbeat interval to be set")
	}

	if config.MessageTimeout == 0 {
		t.Error("Expected default message timeout to be set")
	}
}
