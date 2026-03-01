package ui

import (
	"errors"
	"testing"
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrNotInitialized", ErrNotInitialized, "ui: not initialized"},
		{"ErrAlreadyInitialized", ErrAlreadyInitialized, "ui: already initialized"},
		{"ErrWindowNotFound", ErrWindowNotFound, "ui: window not found"},
		{"ErrChatNotFound", ErrChatNotFound, "ui: chat not found"},
		{"ErrContactNotFound", ErrContactNotFound, "ui: contact not found"},
		{"ErrTaskNotFound", ErrTaskNotFound, "ui: task not found"},
		{"ErrCallNotFound", ErrCallNotFound, "ui: call not found"},
		{"ErrInvalidConfig", ErrInvalidConfig, "ui: invalid config"},
		{"ErrInvalidParameter", ErrInvalidParameter, "ui: invalid parameter"},
		{"ErrOperationFailed", ErrOperationFailed, "ui: operation failed"},
		{"ErrResourceNotFound", ErrResourceNotFound, "ui: resource not found"},
		{"ErrPermissionDenied", ErrPermissionDenied, "ui: permission denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestUIError(t *testing.T) {
	originalErr := errors.New("original error")
	uiErr := NewUIError("TEST_ERROR", "test error message", originalErr)

	if uiErr.Code != "TEST_ERROR" {
		t.Errorf("expected TEST_ERROR, got %s", uiErr.Code)
	}
	if uiErr.Message != "test error message" {
		t.Errorf("expected test error message, got %s", uiErr.Message)
	}
	if uiErr.Err != originalErr {
		t.Error("expected original error to be set")
	}

	expectedFull := "TEST_ERROR: test error message: original error"
	if uiErr.Error() != expectedFull {
		t.Errorf("expected %s, got %s", expectedFull, uiErr.Error())
	}

	if uiErr.Unwrap() != originalErr {
		t.Error("expected Unwrap to return original error")
	}
}

func TestUIErrorWithoutCause(t *testing.T) {
	uiErr := NewUIError("TEST_ERROR", "test error message", nil)

	expectedFull := "TEST_ERROR: test error message"
	if uiErr.Error() != expectedFull {
		t.Errorf("expected %s, got %s", expectedFull, uiErr.Error())
	}

	if uiErr.Unwrap() != nil {
		t.Error("expected Unwrap to return nil when no cause")
	}
}
