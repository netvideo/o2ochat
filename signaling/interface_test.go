package signaling

import (
	"testing"
	"time"
)

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected MessageType
	}{
		{"Offer", MessageTypeOffer},
		{"Answer", MessageTypeAnswer},
		{"Candidate", MessageTypeCandidate},
		{"Invite", MessageTypeInvite},
		{"Bye", MessageTypeBye},
		{"Ping", MessageTypePing},
		{"Pong", MessageTypePong},
		{"Register", MessageTypeRegister},
		{"Lookup", MessageTypeLookup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("message type should not be empty")
			}
		})
	}
}

func TestConnectionState(t *testing.T) {
	tests := []struct {
		name     string
		expected ConnectionState
	}{
		{"Disconnected", StateDisconnected},
		{"Connecting", StateConnecting},
		{"Connected", StateConnected},
		{"Reconnecting", StateReconnecting},
		{"Failed", StateFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("connection state should not be empty")
			}
		})
	}
}

func TestSignalingMessage(t *testing.T) {
	msg := &SignalingMessage{
		Type:      MessageTypeOffer,
		From:      "QmPeer123",
		To:        "QmPeer456",
		Data:      nil,
		Timestamp: time.Now(),
		Signature: []byte("signature"),
		Nonce:     "random-nonce",
	}

	if msg.Type == "" {
		t.Error("message type should not be empty")
	}
	if msg.From == "" {
		t.Error("from should not be empty")
	}
	if msg.To == "" {
		t.Error("to should not be empty")
	}
	if msg.Nonce == "" {
		t.Error("nonce should not be empty")
	}
}

func TestSDPInfo(t *testing.T) {
	sdp := &SDPInfo{
		Type: "offer",
		SDP:  "v=0\r\n...",
	}

	if sdp.Type == "" {
		t.Error("type should not be empty")
	}
	if sdp.SDP == "" {
		t.Error("SDP should not be empty")
	}
}

func TestICECandidate(t *testing.T) {
	candidate := &ICECandidate{
		Candidate:     "candidate:1 1 UDP 2130366237 192.168.1.1 54777 typ host",
		SDPMid:        "0",
		SDPMLineIndex: 0,
	}

	if candidate.Candidate == "" {
		t.Error("candidate should not be empty")
	}
}

func TestPeerInfo(t *testing.T) {
	peerInfo := &PeerInfo{
		PeerID:    "QmPeer123",
		IPv6Addrs: []string{"2001:db8::1"},
		IPv4Addrs: []string{"192.168.1.1"},
		PublicKey: []byte("public-key"),
		LastSeen:  time.Now(),
		Online:    true,
	}

	if peerInfo.PeerID == "" {
		t.Error("peer ID should not be empty")
	}
	if len(peerInfo.IPv6Addrs) == 0 && len(peerInfo.IPv4Addrs) == 0 {
		t.Error("at least one address should be provided")
	}
}

func TestServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	if config.MaxConnections <= 0 {
		t.Error("max connections should be positive")
	}
	if config.HeartbeatInterval <= 0 {
		t.Error("heartbeat interval should be positive")
	}
	if config.MessageTimeout <= 0 {
		t.Error("message timeout should be positive")
	}
	if config.Port <= 0 {
		t.Error("port should be positive")
	}
}

		{ErrConnectionFailed, "ErrConnectionFailed"},
		{ErrMessageSendFailed, "ErrMessageSendFailed"},
		{ErrMessageReceiveFailed, "ErrMessageReceiveFailed"},
		{ErrInvalidMessage, "ErrInvalidMessage"},
		{ErrInvalidPeerID, "ErrInvalidPeerID"},
		{ErrPeerNotFound, "ErrPeerNotFound"},
		{ErrPeerOffline, "ErrPeerOffline"},
		{ErrServerNotRunning, "ErrServerNotRunning"},
		{ErrServerAlreadyRunning, "ErrServerAlreadyRunning"},
		{ErrRegistrationFailed, "ErrRegistrationFailed"},
		{ErrUnregistrationFailed, "ErrUnregistrationFailed"},
		{ErrLookupFailed, "ErrLookupFailed"},
		{ErrSignatureInvalid, "ErrSignatureInvalid"},
		{ErrNonceInvalid, "ErrNonceInvalid"},
		{ErrMessageTimeout, "ErrMessageTimeout"},
		{ErrHeartbeatTimeout, "ErrHeartbeatTimeout"},
		{ErrMaxConnections, "ErrMaxConnections"},
		{ErrDHTJoinFailed, "ErrDHTJoinFailed"},
		{ErrDHTPublishFailed, "ErrDHTPublishFailed"},
		{ErrDHTLookupFailed, "ErrDHTLookupFailed"},
		{ErrCompressionFailed, "ErrCompressionFailed"},
		{ErrDecompressionFailed, "ErrDecompressionFailed"},
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

func TestSignalingError(t *testing.T) {
	innerErr := ErrConnectionFailed
	signalingErr := NewSignalingError("CONN_FAILED", "connection failed", innerErr)

	if signalingErr.Code != "CONN_FAILED" {
		t.Errorf("expected code CONN_FAILED, got %s", signalingErr.Code)
	}
	if signalingErr.Message != "connection failed" {
		t.Errorf("expected message 'connection failed', got %s", signalingErr.Message)
	}
	if signalingErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if signalingErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ SignalingClient = nil
	var _ SignalingServer = nil
	var _ DHTSignaling = nil
	var _ MessageHandler = nil
}
