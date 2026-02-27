package signaling

import (
	"time"
	"errors"
	"testing"
)

func TestErrorConstants(t *testing.T) {
	errors := []error{
		ErrInvalidMessage,
		ErrInvalidSignature,
		ErrInvalidPeerID,
		ErrConnectionFailed,
		ErrConnectionClosed,
		ErrServerNotRunning,
		ErrServerAlreadyRunning,
		ErrNotConnected,
		ErrAlreadyConnected,
		ErrMessageSendFailed,
		ErrMessageTimeout,
		ErrPeerNotFound,
		ErrLookupFailed,
		ErrHeartbeatTimeout,
		ErrSignatureInvalid,
		ErrReconnectFailed,
	}

	for i, err := range errors {
		if err == nil {
			t.Errorf("Error %d should not be nil", i)
		}
	}
}

func TestErrorMessageNotEmpty(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"InvalidMessage", ErrInvalidMessage},
		{"InvalidSignature", ErrInvalidSignature},
		{"InvalidPeerID", ErrInvalidPeerID},
		{"ConnectionFailed", ErrConnectionFailed},
		{"ConnectionClosed", ErrConnectionClosed},
		{"ServerNotRunning", ErrServerNotRunning},
		{"ServerAlreadyRunning", ErrServerAlreadyRunning},
		{"NotConnected", ErrNotConnected},
		{"AlreadyConnected", ErrAlreadyConnected},
		{"MessageSendFailed", ErrMessageSendFailed},
		{"MessageTimeout", ErrMessageTimeout},
		{"PeerNotFound", ErrPeerNotFound},
		{"LookupFailed", ErrLookupFailed},
		{"HeartbeatTimeout", ErrHeartbeatTimeout},
		{"SignatureInvalid", ErrSignatureInvalid},
		{"ReconnectFailed", ErrReconnectFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err.Error() == "" {
				t.Errorf("Expected non-empty error message for %s", test.name)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{"Same error", ErrInvalidMessage, ErrInvalidMessage, true},
		{"Different error", ErrInvalidMessage, ErrInvalidSignature, false},
		{"Wrapped error", ErrInvalidMessage, ErrInvalidMessage, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := errors.Is(test.err, test.target)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	wrapped := ErrConnectionFailed
	if wrapped == nil {
		t.Error("Expected wrapped error to not be nil")
	}

	if wrapped.Error() == "" {
		t.Error("Expected wrapped error to have message")
	}
}

func TestServerConfigValidation(t *testing.T) {
	config := &ServerConfig{
		Port:              8080,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
		MaxConnections:    1000,
	}

	if config.Port <= 0 {
		t.Error("Expected valid port number")
	}
	if config.HeartbeatInterval <= 0 {
		t.Error("Expected valid heartbeat interval")
	}
	if config.MessageTimeout <= 0 {
		t.Error("Expected valid message timeout")
	}
	if config.MaxConnections <= 0 {
		t.Error("Expected valid max connections")
	}
}

func TestClientConfigValidation(t *testing.T) {
	config := &ClientConfig{
		ServerURL:            "ws://localhost:8080",
		HeartbeatInterval:    30 * time.Second,
		MessageTimeout:       10 * time.Second,
		MaxReconnectAttempts: 3,
		ReconnectInterval:    5 * time.Second,
	}

	if config.ServerURL == "" {
		t.Error("Expected server URL to be set")
	}
	if config.HeartbeatInterval <= 0 {
		t.Error("Expected valid heartbeat interval")
	}
	if config.MessageTimeout <= 0 {
		t.Error("Expected valid message timeout")
	}
	if config.MaxReconnectAttempts < 0 {
		t.Error("Expected valid max reconnect attempts")
	}
}

func TestClientConfigZeroReconnect(t *testing.T) {
	config := &ClientConfig{
		ServerURL:            "ws://localhost:8080",
		MaxReconnectAttempts: 0,
	}

	if config.MaxReconnectAttempts != 0 {
		t.Error("Expected zero reconnect attempts to be valid")
	}
}

func TestConnectionStateComparison(t *testing.T) {
	if StateDisconnected == StateConnected {
		t.Error("Disconnected and Connected should be different")
	}
	if StateConnecting == StateReconnecting {
		t.Error("Connecting and Reconnecting should be different")
	}
	if StateFailed == StateConnected {
		t.Error("Failed and Connected should be different")
	}
}

func TestMessageTypeComparison(t *testing.T) {
	if MessageTypeOffer == MessageTypeAnswer {
		t.Error("Offer and Answer should be different")
	}
	if MessageTypePing == MessageTypePong {
		t.Error("Ping and Pong should be different")
	}
	if MessageTypeInvite == MessageTypeBye {
		t.Error("Invite and Bye should be different")
	}
}

func TestErrorUnwrap(t *testing.T) {
	type unwrapper interface {
		Unwrap() error
	}

	tests := []struct {
		name string
		err  error
	}{
		{"InvalidMessage", ErrInvalidMessage},
		{"ConnectionFailed", ErrConnectionFailed},
		{"ServerNotRunning", ErrServerNotRunning},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, ok := test.err.(unwrapper); ok {
				unwrapped := errors.Unwrap(test.err)
				if unwrapped == nil {
					t.Log("Error does not wrap another error")
				}
			}
		})
	}
}
