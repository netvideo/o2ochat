package signaling

import (
	"context"
	"testing"
	"time"
)

func TestSignalingClientInterface(t *testing.T) {
	config := DefaultClientConfig()

	if config.ServerURL == "" {
		t.Error("default server URL should not be empty")
	}
	if config.Timeout <= 0 {
		t.Error("default timeout should be positive")
	}
	if config.MaxRetries < 0 {
		t.Error("max retries should be non-negative")
	}
	if config.ReconnectDelay <= 0 {
		t.Error("reconnect delay should be positive")
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

func TestSignalingErrorCode(t *testing.T) {
	tests := []struct {
		err  SignalingErrorCode
		name string
	}{
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
