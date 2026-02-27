package cli

import "errors"

var (
	ErrCommandNotFound    = errors.New("command not found")
	ErrInvalidArgs        = errors.New("invalid arguments")
	ErrMissingRequiredArg = errors.New("missing required argument")
	ErrInvalidFlag        = errors.New("invalid flag")
	ErrFlagNotFound       = errors.New("flag not found")
	ErrCommandFailed      = errors.New("command execution failed")
	ErrConfigNotFound     = errors.New("config not found")
	ErrConfigInvalid     = errors.New("config invalid")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrFileNotFound      = errors.New("file not found")
	ErrFileReadFailed    = errors.New("file read failed")
	ErrFileWriteFailed   = errors.New("file write failed")
	ErrTimeout           = errors.New("command timeout")
	ErrAlreadyRunning    = errors.New("already running")
	ErrNotRunning        = errors.New("not running")
	ErrInvalidOutputFormat = errors.New("invalid output format")
	ErrCLINotInitialized = errors.New("CLI not initialized")
)

type CLIError struct {
	Code    string
	Message string
	Err     error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

func NewCLIError(code, message string, err error) *CLIError {
	return &CLIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
