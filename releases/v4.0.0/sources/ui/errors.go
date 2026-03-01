package ui

import "errors"

var (
	ErrNotInitialized    = errors.New("ui: not initialized")
	ErrAlreadyInitialized = errors.New("ui: already initialized")
	ErrWindowNotFound    = errors.New("ui: window not found")
	ErrChatNotFound      = errors.New("ui: chat not found")
	ErrContactNotFound   = errors.New("ui: contact not found")
	ErrTaskNotFound      = errors.New("ui: task not found")
	ErrCallNotFound      = errors.New("ui: call not found")
	ErrInvalidConfig     = errors.New("ui: invalid config")
	ErrInvalidParameter  = errors.New("ui: invalid parameter")
	ErrOperationFailed   = errors.New("ui: operation failed")
	ErrResourceNotFound  = errors.New("ui: resource not found")
	ErrPermissionDenied  = errors.New("ui: permission denied")
)

type UIError struct {
	Code    string
	Message string
	Err     error
}

func (e *UIError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *UIError) Unwrap() error {
	return e.Err
}

func NewUIError(code, message string, err error) *UIError {
	return &UIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
