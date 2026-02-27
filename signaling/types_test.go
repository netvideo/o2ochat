package signaling

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageTypeConstants(t *testing.T) {
	if MessageTypeOffer == "" {
		t.Error("MessageTypeOffer should not be empty")
	}
	if MessageTypeAnswer == "" {
		t.Error("MessageTypeAnswer should not be empty")
	}
	if MessageTypeCandidate == "" {
		t.Error("MessageTypeCandidate should not be empty")
	}
	if MessageTypeInvite == "" {
		t.Error("MessageTypeInvite should not be empty")
	}
	if MessageTypeBye == "" {
		t.Error("MessageTypeBye should not be empty")
	}
	if MessageTypePing == "" {
		t.Error("MessageTypePing should not be empty")
	}
	if MessageTypePong == "" {
		t.Error("MessageTypePong should not be empty")
	}
}

func TestSignalingMessageSerialization(t *testing.T) {
	msg := &SignalingMessage{
		Type:      MessageTypeOffer,
		From:      "peer1",
		To:        "peer2",
		Data:      map[string]interface{}{"sdp": "test-sdp"},
		Timestamp: time.Now(),
		Signature: []byte("test-signature"),
		Nonce:     "test-nonce",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty marshaled data")
	}

	var unmarshaled SignalingMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Errorf("Expected type %s, got %s", msg.Type, unmarshaled.Type)
	}
	if unmarshaled.From != msg.From {
		t.Errorf("Expected from %s, got %s", msg.From, unmarshaled.From)
	}
	if unmarshaled.To != msg.To {
		t.Errorf("Expected to %s, got %s", msg.To, unmarshaled.To)
	}
	if unmarshaled.Nonce != msg.Nonce {
		t.Errorf("Expected nonce %s, got %s", msg.Nonce, unmarshaled.Nonce)
	}
}

func TestSignalingMessageEmpty(t *testing.T) {
	msg := &SignalingMessage{}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal empty message: %v", err)
	}

	var unmarshaled SignalingMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty message: %v", err)
	}
}

func TestSDPInfoSerialization(t *testing.T) {
	sdp := SDPInfo{
		Type: "offer",
		SDP:  "v=0\r\no=- 0 0 IN IP4 127.0.0.1",
	}

	data, err := json.Marshal(sdp)
	if err != nil {
		t.Fatalf("Failed to marshal SDP: %v", err)
	}

	var unmarshaled SDPInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SDP: %v", err)
	}

	if unmarshaled.Type != sdp.Type {
		t.Errorf("Expected type %s, got %s", sdp.Type, unmarshaled.Type)
	}
	if unmarshaled.SDP != sdp.SDP {
		t.Errorf("Expected SDP %s, got %s", sdp.SDP, unmarshaled.SDP)
	}
}

func TestICECandidateSerialization(t *testing.T) {
	candidate := ICECandidate{
		Candidate:     "candidate:1 1 UDP 1686052607 192.168.1.1 5000 typ host",
		SDPMid:        "0",
		SDPMLineIndex: 0,
	}

	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatalf("Failed to marshal candidate: %v", err)
	}

	var unmarshaled ICECandidate
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal candidate: %v", err)
	}

	if unmarshaled.Candidate != candidate.Candidate {
		t.Errorf("Expected candidate %s, got %s", candidate.Candidate, unmarshaled.Candidate)
	}
	if unmarshaled.SDPMid != candidate.SDPMid {
		t.Errorf("Expected sdpMid %s, got %s", candidate.SDPMid, unmarshaled.SDPMid)
	}
	if unmarshaled.SDPMLineIndex != candidate.SDPMLineIndex {
		t.Errorf("Expected sdpMLineIndex %d, got %d", candidate.SDPMLineIndex, unmarshaled.SDPMLineIndex)
	}
}

func TestPeerInfoSerialization(t *testing.T) {
	peerInfo := PeerInfo{
		PeerID:    "QmPeer123",
		IPv6Addrs: []string{"2001:db8::1", "2001:db8::2"},
		IPv4Addrs: []string{"192.168.1.1"},
		PublicKey: []byte("test-public-key"),
		LastSeen:  time.Now(),
		Online:    true,
	}

	data, err := json.Marshal(peerInfo)
	if err != nil {
		t.Fatalf("Failed to marshal peer info: %v", err)
	}

	var unmarshaled PeerInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal peer info: %v", err)
	}

	if unmarshaled.PeerID != peerInfo.PeerID {
		t.Errorf("Expected peer ID %s, got %s", peerInfo.PeerID, unmarshaled.PeerID)
	}
	if len(unmarshaled.IPv6Addrs) != len(peerInfo.IPv6Addrs) {
		t.Errorf("Expected %d IPv6 addresses, got %d", len(peerInfo.IPv6Addrs), len(unmarshaled.IPv6Addrs))
	}
	if len(unmarshaled.IPv4Addrs) != len(peerInfo.IPv4Addrs) {
		t.Errorf("Expected %d IPv4 addresses, got %d", len(peerInfo.IPv4Addrs), len(unmarshaled.IPv4Addrs))
	}
	if unmarshaled.Online != peerInfo.Online {
		t.Errorf("Expected online %v, got %v", peerInfo.Online, unmarshaled.Online)
	}
}

func TestPeerInfoEmpty(t *testing.T) {
	peerInfo := PeerInfo{}

	data, err := json.Marshal(peerInfo)
	if err != nil {
		t.Fatalf("Failed to marshal empty peer info: %v", err)
	}

	var unmarshaled PeerInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty peer info: %v", err)
	}
}

func TestPeerInfoWithNilAddresses(t *testing.T) {
	peerInfo := PeerInfo{
		PeerID: "QmPeer123",
		Online: true,
	}

	data, err := json.Marshal(peerInfo)
	if err != nil {
		t.Fatalf("Failed to marshal peer info with nil addresses: %v", err)
	}

	var unmarshaled PeerInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal peer info: %v", err)
	}

	if unmarshaled.PeerID != peerInfo.PeerID {
		t.Errorf("Expected peer ID %s, got %s", peerInfo.PeerID, unmarshaled.PeerID)
	}
}

func TestConnectionStateConstants(t *testing.T) {
	states := []ConnectionState{
		StateDisconnected,
		StateConnecting,
		StateConnected,
		StateReconnecting,
		StateFailed,
	}

	for i, state := range states {
		if state == "" {
			t.Errorf("Connection state %d should not be empty", i)
		}
	}
}

func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateDisconnected, "disconnected"},
		{StateConnecting, "connecting"},
		{StateConnected, "connected"},
		{StateReconnecting, "reconnecting"},
		{StateFailed, "failed"},
	}

	for _, test := range tests {
		if test.state.String() != test.expected {
			t.Errorf("Expected state string %s, got %s", test.expected, test.state.String())
		}
	}
}

func TestServerConfigDefaults(t *testing.T) {
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
	if config.MaxConnections == 0 {
		t.Error("Expected default max connections to be set")
	}
}

func TestClientConfigDefaults(t *testing.T) {
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
	if config.MaxReconnectAttempts == 0 {
		t.Error("Expected default max reconnect attempts to be set")
	}
}

func TestServerConfigCopy(t *testing.T) {
	config1 := &ServerConfig{
		Port:              8080,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
	}

	config2 := *config1

	if config2.Port != config1.Port {
		t.Error("Expected port to be copied")
	}

	config2.Port = 9090
	if config1.Port == 9090 {
		t.Error("Expected config1 to remain unchanged")
	}
}

func TestClientConfigCopy(t *testing.T) {
	config1 := &ClientConfig{
		ServerURL:         "ws://localhost:8080",
		HeartbeatInterval: 30 * time.Second,
	}

	config2 := *config1

	if config2.ServerURL != config1.ServerURL {
		t.Error("Expected server URL to be copied")
	}

	config2.ServerURL = "ws://localhost:9090"
	if config1.ServerURL == "ws://localhost:9090" {
		t.Error("Expected config1 to remain unchanged")
	}
}

func TestMessageWithComplexData(t *testing.T) {
	complexData := map[string]interface{}{
		"sdp": map[string]interface{}{
			"type": "offer",
			"sdp":  "v=0\r\n",
		},
		"candidates": []interface{}{
			map[string]interface{}{"candidate": "c1"},
			map[string]interface{}{"candidate": "c2"},
		},
	}

	msg := &SignalingMessage{
		Type: MessageTypeOffer,
		From: "peer1",
		To:   "peer2",
		Data: complexData,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message with complex data: %v", err)
	}

	var unmarshaled SignalingMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal message with complex data: %v", err)
	}

	if unmarshaled.Data == nil {
		t.Error("Expected data to be preserved")
	}
}
