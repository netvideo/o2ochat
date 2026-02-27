package transport

import (
	"testing"
	"time"
)

func TestConnectionTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected ConnectionType
	}{
		{"QUIC", ConnectionTypeQUIC},
		{"WebRTC", ConnectionTypeWebRTC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("connection type should not be empty")
			}
		})
	}
}

func TestConnectionStates(t *testing.T) {
	tests := []struct {
		name     string
		expected ConnectionState
	}{
		{"Disconnected", StateDisconnected},
		{"Connecting", StateConnecting},
		{"Connected", StateConnected},
		{"Failed", StateFailed},
		{"Closing", StateClosing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("connection state should not be empty")
			}
		})
	}
}

func TestNetworkTypes(t *testing.T) {
	tests := []struct {
		name     string
		expected NetworkType
	}{
		{"IPv6", NetworkTypeIPv6},
		{"IPv4", NetworkTypeIPv4},
		{"Unknown", NetworkTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.expected) == "" {
				t.Error("network type should not be empty")
			}
		})
	}
}

func TestConnectionConfig(t *testing.T) {
	config := DefaultConnectionConfig()

	if config.Timeout <= 0 {
		t.Error("timeout should be positive")
	}
	if config.RetryCount <= 0 {
		t.Error("retry count should be positive")
	}
	if len(config.Priority) == 0 {
		t.Error("priority should not be empty")
	}
}

func TestConnectionStats(t *testing.T) {
	stats := &ConnectionStats{
		BytesSent:       1000,
		BytesReceived:   500,
		PacketsSent:     10,
		PacketsReceived: 5,
		Retransmits:     0,
		Latency:         50 * time.Millisecond,
		Bandwidth:       1000000,
		PacketLoss:      0.0,
	}

	if stats.BytesSent < 0 {
		t.Error("bytes sent should not be negative")
	}
	if stats.PacketLoss < 0 || stats.PacketLoss > 1 {
		t.Error("packet loss should be between 0 and 1")
	}
}

func TestStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()

	if !config.Reliable {
		t.Error("reliable should be true by default")
	}
	if !config.Ordered {
		t.Error("ordered should be true by default")
	}
	if config.BufferSize <= 0 {
		t.Error("buffer size should be positive")
	}
}

func TestStreamInfo(t *testing.T) {
	info := &StreamInfo{
		ID:        1,
		Direction: "bidirectional",
		State:     "open",
	}

	if info.ID == 0 {
		t.Error("stream ID should not be zero")
	}
	if info.Direction == "" {
		t.Error("direction should not be empty")
	}
	if info.State == "" {
		t.Error("state should not be empty")
	}
}

func TestQUICConfig(t *testing.T) {
	config := DefaultQUICConfig()

	if config.MaxIncomingStreams <= 0 {
		t.Error("max incoming streams should be positive")
	}
	if config.MaxIdleTimeout <= 0 {
		t.Error("max idle timeout should be positive")
	}
}

func TestWebRTCConfig(t *testing.T) {
	config := DefaultWebRTCConfig()

	if len(config.ICEServers) == 0 {
		t.Error("ICE servers should not be empty")
	}
	if config.PortRange.Min <= 0 || config.PortRange.Max <= 0 {
		t.Error("port range should be positive")
	}
	if config.PortRange.Min > config.PortRange.Max {
		t.Error("min port should be less than max port")
	}
}

func TestICEServer(t *testing.T) {
	server := ICEServer{
		URLs:           []string{"stun:stun.l.google.com:19302"},
		Username:       "user",
		Credential:     "pass",
		CredentialType: "password",
	}

	if len(server.URLs) == 0 {
		t.Error("URLs should not be empty")
	}
}

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{ErrNotConnected, "ErrNotConnected"},
		{ErrAlreadyConnected, "ErrAlreadyConnected"},
		{ErrConnectionFailed, "ErrConnectionFailed"},
		{ErrConnectionClosed, "ErrConnectionClosed"},
		{ErrNoAvailableAddress, "ErrNoAvailableAddress"},
		{ErrHandshakeTimeout, "ErrHandshakeTimeout"},
		{ErrStreamOpenFailed, "ErrStreamOpenFailed"},
		{ErrStreamClosed, "ErrStreamClosed"},
		{ErrWriteTimeout, "ErrWriteTimeout"},
		{ErrReadTimeout, "ErrReadTimeout"},
		{ErrInvalidConfig, "ErrInvalidConfig"},
		{ErrInvalidAddress, "ErrInvalidAddress"},
		{ErrUnsupportedProtocol, "ErrUnsupportedProtocol"},
		{ErrNATTraversalFailed, "ErrNATTraversalFailed"},
		{ErrHolePunchingFailed, "ErrHolePunchingFailed"},
		{ErrRelayFailed, "ErrRelayFailed"},
		{ErrListenerClosed, "ErrListenerClosed"},
		{ErrMaxStreamsReached, "ErrMaxStreamsReached"},
		{ErrBandwidthExceeded, "ErrBandwidthExceeded"},
		{ErrNetworkChanged, "ErrNetworkChanged"},
		{ErrEncryptionFailed, "ErrEncryptionFailed"},
		{ErrPeerNotReachable, "ErrPeerNotReachable"},
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

func TestTransportError(t *testing.T) {
	innerErr := ErrConnectionFailed
	transportErr := NewTransportError("CONN_FAILED", "connection failed", innerErr)

	if transportErr.Code != "CONN_FAILED" {
		t.Errorf("expected code CONN_FAILED, got %s", transportErr.Code)
	}
	if transportErr.Message != "connection failed" {
		t.Errorf("expected message 'connection failed', got %s", transportErr.Message)
	}
	if transportErr.Unwrap() != innerErr {
		t.Error("unwrap should return inner error")
	}
	if transportErr.Error() == "" {
		t.Error("error should not be empty")
	}
}

func TestInterfaceCompatibility(t *testing.T) {
	var _ TransportManager = nil
	var _ Connection = nil
	var _ Stream = nil
	var _ NATTraversal = nil
	var _ ConnectionHandler = nil
	var _ StreamHandler = nil
}
