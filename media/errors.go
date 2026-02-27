package media

import "errors"

var (
	ErrNotInitialized     = errors.New("media manager not initialized")
	ErrDeviceNotFound     = errors.New("device not found")
	ErrDeviceNotAvailable = errors.New("device not available")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrCodecNotSupported  = errors.New("codec not supported")
	ErrEncodingFailed    = errors.New("encoding failed")
	ErrDecodingFailed    = errors.New("decoding failed")
	ErrFrameTooLarge     = errors.New("frame too large")
	ErrInvalidFrame       = errors.New("invalid frame")
	ErrRTPPacketizeFailed = errors.New("RTP packetize failed")
	ErrRTPDepacketizeFailed = errors.New("RTP depacketize failed")
	ErrBufferFull         = errors.New("buffer full")
	ErrBufferEmpty        = errors.New("buffer empty")
	ErrCallNotActive     = errors.New("call not active")
	ErrCallAlreadyActive = errors.New("call already active")
	ErrBitrateAdjustFailed = errors.New("bitrate adjust failed")
	ErrDeviceSwitchFailed = errors.New("device switch failed")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrNoAudioDevice     = errors.New("no audio device")
	ErrNoVideoDevice     = errors.New("no video device")
)

type MediaError struct {
	Code    string
	Message string
	Err     error
}

func (e *MediaError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *MediaError) Unwrap() error {
	return e.Err
}

func NewMediaError(code, message string, err error) *MediaError {
	return &MediaError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
